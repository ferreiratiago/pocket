package e2e_tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/logger"
	mocksPer "github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// TODO(integration): These are temporary variables used in the prototype integration phase that
// will need to be parameterized later once the test framework design matures.
const (
	numValidators          = 4
	stateHash              = "42"
	stateSyncMaxBlockCount = 1000
)

var maxTxBytes = defaults.DefaultConsensusMaxMempoolBytes

type IdToNodeMapping map[typesCons.NodeId]*shared.Node

// CONSIDER: Remove this global variable and pass it around as a parameter.
var stateSyncDummyblocks []*coreTypes.Block

/*** Node Generation Helpers ***/

func GenerateNodeRuntimeMgrs(_ *testing.T, validatorCount int, clockMgr clock.Clock) []*runtime.Manager {
	runtimeMgrs := make([]*runtime.Manager, validatorCount)
	var validatorKeys []string
	genesisState, validatorKeys := test_artifacts.NewGenesisState(validatorCount, 1, 1, 1)
	cfgs := test_artifacts.NewDefaultConfigs(validatorKeys)
	for i, config := range cfgs {
		config.Consensus = &configs.ConsensusConfig{
			PrivateKey:      config.PrivateKey,
			MaxMempoolBytes: maxTxBytes,
			PacemakerConfig: &configs.PacemakerConfig{
				TimeoutMsec:               10000,
				Manual:                    false,
				DebugTimeBetweenStepsMsec: 0,
			},
			ServerModeEnabled: true,
		}
		runtimeMgrs[i] = runtime.NewManager(config, genesisState, runtime.WithClock(clockMgr))
	}
	return runtimeMgrs
}

func CreateTestConsensusPocketNodes(
	t *testing.T,
	buses []modules.Bus,
	eventsChannel modules.EventsChannel,
) (pocketNodes IdToNodeMapping) {
	pocketNodes = make(IdToNodeMapping, len(buses))
	// TODO(design): The order here is important in order for NodeId to be set correctly below.
	// This logic will need to change once proper leader election is implemented.
	sort.Slice(buses, func(i, j int) bool {
		pk, err := cryptoPocket.NewPrivateKey(buses[i].GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)
		pk2, err := cryptoPocket.NewPrivateKey(buses[j].GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)
		return pk.Address().String() < pk2.Address().String()
	})

	for i := range buses {
		pocketNode := CreateTestConsensusPocketNode(t, buses[i], eventsChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked
func CreateTestConsensusPocketNode(
	t *testing.T,
	bus modules.Bus,
	eventsChannel modules.EventsChannel,
) *shared.Node {
	persistenceMock := basePersistenceMock(t, eventsChannel, bus)
	bus.RegisterModule(persistenceMock)

	consensusMod, err := consensus.Create(bus)
	require.NoError(t, err)
	consensusModule, ok := consensusMod.(modules.ConsensusModule)
	require.True(t, ok)

	runtimeMgr := (bus).GetRuntimeMgr()
	// TODO(olshansky): At the moment we are using the same base mocks for all the tests,
	// but note that they will need to be customized on a per test basis.
	p2pMock := baseP2PMock(t, eventsChannel)
	utilityMock := baseUtilityMock(t, eventsChannel, runtimeMgr.GetGenesis(), consensusModule)
	telemetryMock := baseTelemetryMock(t, eventsChannel)
	loggerMock := baseLoggerMock(t, eventsChannel)
	rpcMock := baseRpcMock(t, eventsChannel)
	stateMachineMock := baseStateMachineMock(t, eventsChannel, bus)

	for _, module := range []modules.Module{
		p2pMock,
		utilityMock,
		telemetryMock,
		loggerMock,
		rpcMock,
		stateMachineMock,
	} {
		bus.RegisterModule(module)
	}

	require.NoError(t, err)

	pk, err := cryptoPocket.NewPrivateKey(runtimeMgr.GetConfig().PrivateKey)
	require.NoError(t, err)

	pocketNode := shared.NewNodeWithP2PAddress(pk.Address())

	pocketNode.SetBus(bus)

	return pocketNode
}

func GenerateBuses(t *testing.T, runtimeMgrs []*runtime.Manager) (buses []modules.Bus) {
	buses = make([]modules.Bus, len(runtimeMgrs))
	for i := range runtimeMgrs {
		bus, err := runtime.CreateBus(runtimeMgrs[i])
		require.NoError(t, err)
		buses[i] = bus
	}
	return
}

// CLEANUP: Reduce package scope visibility in the consensus test module
func StartAllTestPocketNodes(t *testing.T, pocketNodes IdToNodeMapping) {
	for _, pocketNode := range pocketNodes {
		go startNode(t, pocketNode)
		startEvent := pocketNode.GetBus().GetBusEvent()
		require.Equal(t, startEvent.GetContentType(), messaging.NodeStartedEventType)
	}
}

/*** Node Visibility/Reflection Helpers ***/

// TODO(discuss): Should we use reflections inside the testing module as being done here or explicitly
// define the interfaces used for debug/development. The latter will probably scale more but will
// require more effort and pollute the source code with debugging information.
func GetConsensusNodeState(node *shared.Node) typesCons.ConsensusNodeState {
	return GetConsensusModImpl(node).MethodByName("GetNodeState").Call([]reflect.Value{})[0].Interface().(typesCons.ConsensusNodeState)
}

func GetConsensusModElem(node *shared.Node) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).Elem()
}

func GetConsensusModImpl(node *shared.Node) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule())
}

/*** Debug/Development Message Helpers ***/

func TriggerNextView(t *testing.T, node *shared.Node) {
	triggerDebugMessage(t, node, messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW)
}

func triggerDebugMessage(t *testing.T, node *shared.Node, action messaging.DebugMessageAction) {
	debugMessage := &messaging.DebugMessage{
		Action:  action,
		Message: nil,
	}
	anyProto, err := anypb.New(debugMessage)
	require.NoError(t, err)

	e := &messaging.PocketEnvelope{Content: anyProto}
	node.GetBus().PublishEventToBus(e)
}

/*** P2P Helpers ***/

func P2PBroadcast(_ *testing.T, nodes IdToNodeMapping, any *anypb.Any) {
	e := &messaging.PocketEnvelope{Content: any}
	for _, node := range nodes {
		node.GetBus().PublishEventToBus(e)
	}
}

func P2PSend(_ *testing.T, node *shared.Node, any *anypb.Any) {
	e := &messaging.PocketEnvelope{Content: any}
	node.GetBus().PublishEventToBus(e)
}

// This is a helper for `waitForEventsInternal` that creates the `includeFilter` function based on
// consensus specific parameters.
// failOnExtraMessages:
//
//	This flag is useful when running the consensus unit tests. It causes the test to wait up to the
//	maximum delay specified in the source code and errors if additional unexpected messages are received.
//	For example, if the test expects to receive 5 messages within 2 seconds:
//		false: continue if 5 messages are received in 0.5 seconds
//		true: wait for another 1.5 seconds after 5 messages are received in 0.5 seconds, and fail if any additional messages are received.
func WaitForNetworkConsensusEvents(
	t *testing.T,
	clck *clock.Mock,
	eventsChannel modules.EventsChannel,
	step typesCons.HotstuffStep,
	msgType typesCons.HotstuffMessageType,
	numExpectedMsgs int,
	millis time.Duration,
	failOnExtraMessages bool,
) (messages []*anypb.Any, err error) {
	includeFilter := func(anyMsg *anypb.Any) bool {
		msg, err := codec.GetCodec().FromAny(anyMsg)
		require.NoError(t, err)

		hotstuffMessage, ok := msg.(*typesCons.HotstuffMessage)
		require.True(t, ok)

		return hotstuffMessage.Type == msgType && hotstuffMessage.Step == step
	}

	errMsg := fmt.Sprintf("HotStuff step: %s, type: %s", typesCons.HotstuffStep_name[int32(step)], typesCons.HotstuffMessageType_name[int32(msgType)])
	return waitForEventsInternal(clck, eventsChannel, messaging.HotstuffMessageContentType, numExpectedMsgs, millis, includeFilter, errMsg, failOnExtraMessages)
}

// IMPROVE: Consider unifying this function with WaitForNetworkConsensusEvents
// This is a helper for 'waitForEventsInternal' that creates the `includeFilter` function based on state sync message specific parameters.
func WaitForNetworkStateSyncEvents(
	t *testing.T,
	clck *clock.Mock,
	eventsChannel modules.EventsChannel,
	errMsg string,
	numExpectedMsgs int,
	millis time.Duration,
	failOnExtraMessages bool,
) (messages []*anypb.Any, err error) {
	includeFilter := func(anyMsg *anypb.Any) bool {
		msg, err := codec.GetCodec().FromAny(anyMsg)
		require.NoError(t, err)

		_, ok := msg.(*typesCons.StateSyncMessage)
		require.True(t, ok)

		return true
	}

	return waitForEventsInternal(clck, eventsChannel, messaging.StateSyncMessageContentType, numExpectedMsgs, millis, includeFilter, errMsg, failOnExtraMessages)
}

// RESEARCH(#462): Research ways to eliminate time-based non-determinism from the test framework
// IMPROVE: This function can be extended to testing events outside of just the consensus module.
func waitForEventsInternal(
	clck *clock.Mock,
	eventsChannel modules.EventsChannel,
	eventContentType string,
	numExpectedMsgs int,
	maxWaitTime time.Duration,
	msgIncludeFilter func(m *anypb.Any) bool,
	errMsg string,
	failOnExtraMessages bool,
) (expectedMsgs []*anypb.Any, err error) {
	expectedMsgs = make([]*anypb.Any, 0)                 // Aggregate and return the messages we're waiting for
	unusedEvents := make([]*messaging.PocketEnvelope, 0) // "Recycle" events back into the events channel if we're not using them

	// Limit the amount of time we're waiting for the messages to be published on the events channel
	ctx, cancel := clck.WithTimeout(context.TODO(), time.Millisecond*maxWaitTime)
	defer cancel()

	// Since the tests use a mock clock, we need to manually advance the clock to trigger the timeout
	ticker := time.NewTicker(time.Millisecond)
	tickerDone := make(chan bool)
	go func() {
		for {
			select {
			case <-tickerDone:
				return
			case <-ticker.C:
				clck.Add(time.Millisecond)
			}
		}
	}()
	defer ticker.Stop()
	defer func() {
		tickerDone <- true
	}()

	numRemainingMsgs := numExpectedMsgs
loop:
	for {
		select {
		case nodeEvent := <-eventsChannel:
			if nodeEvent.GetContentType() != eventContentType {
				unusedEvents = append(unusedEvents, nodeEvent)
				continue
			}

			message := nodeEvent.Content
			if message == nil || !msgIncludeFilter(message) {
				unusedEvents = append(unusedEvents, nodeEvent)
				continue
			}

			expectedMsgs = append(expectedMsgs, message)
			numRemainingMsgs--
			// Break if both of the following are true:
			// 1. We are not expecting any more messages
			// 2. We do not want to fail in the case of extra unexpected messages that pass the filter
			if numRemainingMsgs == 0 && !failOnExtraMessages {
				break loop
			}
		// The reason we return we format and return an error message rather than using t.Fail(...)
		// is to allow the called to run `require.NoError(t, err)` and have the output point to the
		// specific line where the test failed.
		case <-ctx.Done():
			if numRemainingMsgs == 0 {
				break loop
			} else if numRemainingMsgs > 0 {
				return expectedMsgs, fmt.Errorf("Missing '%s' messages; %d expected but %d received. (%s) \n\t DO_NOT_SKIP_ME(#462): Consider increasing `maxWaitTimeMillis` as a workaround", eventContentType, numExpectedMsgs, len(expectedMsgs), errMsg)
			} else {
				return expectedMsgs, fmt.Errorf("Too many '%s' messages; %d expected but %d received. (%s)", eventContentType, numExpectedMsgs, len(expectedMsgs), errMsg)
			}
		}
	}

	for _, u := range unusedEvents {
		eventsChannel <- u
	}
	return
}

/*** Module Mocking Helpers ***/

// Creates a persistence module mock with mock implementations of some basic functionality
func basePersistenceMock(t *testing.T, _ modules.EventsChannel, bus modules.Bus) *mockModules.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	persistenceReadContextMock := mockModules.NewMockPersistenceReadContext(ctrl)

	//persistence block must have access to the dummy valid blocks
	// CHECK IF NUMVALIDATORS CAN BE AN ISSUE HERE
	//pocketNodes
	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()

	persistenceMock.EXPECT().ReleaseWriteContext().Return(nil).AnyTimes()

	blockStoreMock := mocksPer.NewMockKVStore(ctrl)

	blockStoreMock.EXPECT().Get(gomock.Any()).DoAndReturn(func(height []byte) ([]byte, error) {
		heightInt := utils.HeightFromBytes(height)
		if bus.GetConsensusModule().CurrentHeight() < heightInt {
			return nil, fmt.Errorf("requested height is higher than current height of the node's consensus module")
		}

		block := stateSyncDummyblocks[heightInt-1]
		return codec.GetCodec().Marshal(block)
	}).AnyTimes()

	persistenceMock.EXPECT().GetBlockStore().Return(blockStoreMock).AnyTimes()

	persistenceReadContextMock.EXPECT().GetMaximumBlockHeight().DoAndReturn(func() (uint64, error) {
		currentHeight := bus.GetConsensusModule().CurrentHeight()
		var maxPersistedHeight uint64
		if currentHeight == 0 {
			maxPersistedHeight = currentHeight
		} else {
			maxPersistedHeight = currentHeight - 1
		}

		return maxPersistedHeight, nil
	}).AnyTimes()

	persistenceReadContextMock.EXPECT().GetMinimumBlockHeight().DoAndReturn(func() (uint64, error) {
		// mock minimum block height in persistence module to 1 if current height is equal or more than 1, else return 0 as the minimum height
		if bus.GetConsensusModule().CurrentHeight() >= 1 {
			return 1, nil
		}
		return 0, nil
	}).AnyTimes()

	persistenceReadContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(bus.GetRuntimeMgr().GetGenesis().Validators, nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()
	persistenceReadContextMock.EXPECT().Release().AnyTimes()

	return persistenceMock
}

// Creates a p2p module mock with mock implementations of some basic functionality
func baseP2PMock(t *testing.T, eventsChannel modules.EventsChannel) *mockModules.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := mockModules.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			eventsChannel <- e
		}).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			eventsChannel <- e
		}).
		AnyTimes()
	p2pMock.EXPECT().GetModuleName().Return(modules.P2PModuleName).AnyTimes()
	// TODO! Update, send event p2p bootstrapped depending on type
	p2pMock.EXPECT().HandleEvent(gomock.Any()).Return(nil).AnyTimes()

	return p2pMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(t *testing.T, _ modules.EventsChannel, genesisState *genesis.GenesisState, consensusMod modules.ConsensusModule) *mockModules.MockUtilityModule {
	ctrl := gomock.NewController(t)
	utilityMock := mockModules.NewMockUtilityModule(ctrl)
	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	utilityMock.EXPECT().
		// NewContext(gomock.Any()).
		// Return(utilityContextMock, nil).
		// MaxTimes(stateSyncMaxBlockCount)
		// utilityMock.EXPECT().
		// 	NewContext(gomock.Any()).
		// 	Return(utilityContextMock, nil)
		NewUnitOfWork(gomock.Any()).
		DoAndReturn(
			// mimicking the behavior of the utility module's NewUnitOfWork method
			func(height int64) (modules.UtilityUnitOfWork, error) {
				if consensusMod.IsLeader() {
					return baseLeaderUtilityUnitOfWorkMock(t, genesisState), nil
				}
				return baseReplicaUtilityUnitOfWorkMock(t, genesisState), nil
			}).
		MaxTimes(stateSyncMaxBlockCount)

	utilityMock.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()

	return utilityMock
}

func baseLeaderUtilityUnitOfWorkMock(t *testing.T, genesisState *genesis.GenesisState) *mockModules.MockLeaderUtilityUnitOfWork {
	ctrl := gomock.NewController(t)
	utilityLeaderUnitOfWorkMock := mockModules.NewMockLeaderUtilityUnitOfWork(ctrl)

	rwContextMock := mockModules.NewMockPersistenceRWContext(ctrl)
	rwContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	rwContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()
	rwContextMock.EXPECT().Release().AnyTimes()

	utilityLeaderUnitOfWorkMock.EXPECT().
		CreateAndApplyProposalBlock(gomock.Any(), maxTxBytes).
		Return(stateHash, make([][]byte, 0), nil).
		AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().
		ApplyBlock().
		Return(stateHash, make([][]byte, 0), nil).
		AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().SetProposalBlock(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().Release().Return(nil).AnyTimes()

	return utilityLeaderUnitOfWorkMock
}

func baseReplicaUtilityUnitOfWorkMock(t *testing.T, genesisState *genesis.GenesisState) *mockModules.MockReplicaUtilityUnitOfWork {
	ctrl := gomock.NewController(t)
	utilityReplicaUnitOfWorkMock := mockModules.NewMockReplicaUtilityUnitOfWork(ctrl)

	rwContextMock := mockModules.NewMockPersistenceRWContext(ctrl)
	rwContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	rwContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()
	rwContextMock.EXPECT().Release().AnyTimes()

	utilityReplicaUnitOfWorkMock.EXPECT().
		ApplyBlock().
		Return(stateHash, make([][]byte, 0), nil).
		AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().SetProposalBlock(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().Release().Return(nil).AnyTimes()

	return utilityReplicaUnitOfWorkMock
}

func baseTelemetryMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := mockModules.NewMockTelemetryModule(ctrl)
	timeSeriesAgentMock := baseTelemetryTimeSeriesAgentMock(t)
	eventMetricsAgentMock := baseTelemetryEventMetricsAgentMock(t)

	telemetryMock.EXPECT().Start().Return(nil).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()

	return telemetryMock
}

func baseRpcMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockRPCModule {
	ctrl := gomock.NewController(t)
	rpcMock := mockModules.NewMockRPCModule(ctrl)
	rpcMock.EXPECT().Start().Return(nil).AnyTimes()
	rpcMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	rpcMock.EXPECT().GetModuleName().Return(modules.RPCModuleName).AnyTimes()

	return rpcMock
}

func baseStateMachineMock(t *testing.T, _ modules.EventsChannel, bus modules.Bus) *mockModules.MockStateMachineModule {
	ctrl := gomock.NewController(t)
	stateMachineMock := mockModules.NewMockStateMachineModule(ctrl)
	stateMachineMock.EXPECT().Start().Return(nil).AnyTimes()
	stateMachineMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	stateMachineMock.EXPECT().GetModuleName().Return(modules.StateMachineModuleName).AnyTimes()

	//consensusModImpl := reflect.ValueOf(bus.GetConsensusModule())

	stateMachineMock.EXPECT().SendEvent(gomock.Any()).DoAndReturn(func(event coreTypes.StateMachineEvent, args ...any) error {
		switch coreTypes.StateMachineEvent(event) {
		case coreTypes.StateMachineEvent_Consensus_IsUnsynched:
			t.Logf("Node is unsynched")
			return bus.GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncing)
		case coreTypes.StateMachineEvent_Consensus_IsSyncing:
			//maxHeight := bus.GetConsensusModule().GetAggregatedStateSyncMetadataMaxHeight()
			t.Logf("CALLING Node is syncing")
			//consensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(maxHeight)})
			//consensusModImpl.MethodByName("TriggerSync").Call([]reflect.Value{})
			bus.GetConsensusModule().DebugTriggerSync()

			//bus.GetConsensusModule().TriggerSync()
			return nil
		case coreTypes.StateMachineEvent_Consensus_IsSynchedValidator:
			t.Logf("Validator node is synched")
			return nil
		case coreTypes.StateMachineEvent_Consensus_IsSynchedNonValidator:
			t.Logf("Non-validator node is synched")
			return nil
		default:
			log.Printf("Not handling this event: %s", event)
			return nil
		}
	}).AnyTimes()

	return stateMachineMock
}

func baseTelemetryTimeSeriesAgentMock(t *testing.T) *mockModules.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeSeriesAgentMock := mockModules.NewMockTimeSeriesAgent(ctrl)
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).MaxTimes(1)
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeSeriesAgentMock
}

func baseTelemetryEventMetricsAgentMock(t *testing.T) *mockModules.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mockModules.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}

func baseLoggerMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockLoggerModule {
	ctrl := gomock.NewController(t)
	loggerMock := mockModules.NewMockLoggerModule(ctrl)

	loggerMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	loggerMock.EXPECT().GetModuleName().Return(modules.LoggerModuleName).AnyTimes()

	return loggerMock
}

func GenerateDummyBlocksWithQC(t *testing.T, numberOfBlocks, numberOfValidators uint64, pocketNodes IdToNodeMapping) { //[]*coreTypes.Block {
	blocks := make([]*coreTypes.Block, 0)

	var i uint64 = 1
	for i <= numberOfBlocks {
		// given height find the leader
		leaderId := i % uint64(numberOfValidators)
		if leaderId == 0 {
			leaderId = numberOfValidators
		}
		leader := pocketNodes[typesCons.NodeId(leaderId)]
		leaderPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
		require.NoError(t, err)

		// Construct the block
		blockHeader := &coreTypes.BlockHeader{
			Height:            i,
			StateHash:         stateHash,
			PrevStateHash:     stateHash,
			ProposerAddress:   leaderPK.Address(),
			QuorumCertificate: nil,
		}
		block := &coreTypes.Block{
			BlockHeader:  blockHeader,
			Transactions: make([][]byte, 0),
		}

		block, err = generateValidBlockWithQC(pocketNodes, block)
		require.NoError(t, err)

		blocks = append(blocks, block)
		i++
	}

	stateSyncDummyblocks = blocks
}

func generateValidBlockWithQC(pocketNodes IdToNodeMapping, block *coreTypes.Block) (*coreTypes.Block, error) {
	qc, err := generateValidQuorumCertificate(pocketNodes, block)
	if err != nil {
		return nil, err
	}

	qcBytes, err := codec.GetCodec().Marshal(qc)
	if err != nil {
		log.Fatalf("could not marshal quorum certificate to bytes: %v", err)
		return nil, err
	}

	block.BlockHeader.QuorumCertificate = qcBytes

	//block.BlockHeader.QuorumCertificate = nil

	return block, nil

}

func logTime(t *testing.T, clck *clock.Mock) {
	t.Helper()
	defer func() {
		// this is to recover from a panic that could happen if the goroutine tries to log after the test has finished
		// cause of the panic: https://github.com/golang/go/blob/135c470b2277e1c9514ba8a5478408fea0dee8a2/src/testing/testing.go#L1003
		//
		// spotted for the first time in our CI: https://github.com/pokt-network/pocket/actions/runs/4198025819/jobs/7281103860#step:8:1118
		//nolint:errcheck // ignoring completely
		recover()
	}()
	t.Logf("[⌚ CLOCK ⌚] the time is: %v ms from UNIX Epoch [%v]", clck.Now().UTC().UnixMilli(), clck.Now().UTC())
}

// advanceTime moves the time forward on the mock clock and logs what just happened.
func advanceTime(t *testing.T, clck *clock.Mock, duration time.Duration) {
	t.Helper()
	clck.Add(duration)
	t.Logf("[⌚ CLOCK ⏩] advanced by %v", duration)
	logTime(t, clck)
}

// sleep pauses the goroutine for the given duration on the mock clock and logs what just happened.
//
// Note: time has to be moved forward in a separate goroutine, see `advanceTime`.
func sleep(t *testing.T, clck *clock.Mock, duration time.Duration) {
	t.Helper()
	t.Logf("[⌚ CLOCK 💤] sleeping for %v", duration)
	clck.Sleep(duration)
}

// timeReminder simply prints, at a given interval and in a separate goroutine, the current mocked time to help with events.
// nolint:unparam // we want to keep the frequency as a parameter for clarity
func timeReminder(t *testing.T, clck *clock.Mock, frequency time.Duration) {
	go func() {
		tick := time.NewTicker(frequency)
		for {
			<-tick.C
			logTime(t, clck)
		}
	}()
}

func assertNodeConsensusView(t *testing.T, nodeId typesCons.NodeId, expected, actual typesCons.ConsensusNodeState) {
	assertHeight(t, nodeId, expected.Height, actual.Height)
	assertStep(t, nodeId, typesCons.HotstuffStep(expected.Step), typesCons.HotstuffStep(actual.Step))
	assertRound(t, nodeId, expected.Round, actual.Round)
}

func assertHeight(t *testing.T, nodeId typesCons.NodeId, expected, actual uint64) {
	require.Equal(t, expected, actual, "[NODE][%v] failed assertHeight", nodeId)
}

func assertStep(t *testing.T, nodeId typesCons.NodeId, expected, actual typesCons.HotstuffStep) {
	require.Equal(t, expected.String(), actual.String(), "[NODE][%v] failed assertStep", nodeId)
}

func assertRound(t *testing.T, nodeId typesCons.NodeId, expected, actual uint8) {
	require.Equal(t, expected, actual, "[NODE][%v] failed assertRound", nodeId)
}

func startNode(t *testing.T, pocketNode *shared.Node) {
	err := pocketNode.Start()
	require.NoError(t, err)
}

func generateValidQuorumCertificate(pocketNodes IdToNodeMapping, block *coreTypes.Block) (*typesCons.QuorumCertificate, error) {
	var pss []*typesCons.PartialSignature

	for _, val := range pocketNodes {
		ps, err := generatePartialSignature(block, val)
		if err != nil {
			return nil, err
		}
		pss = append(pss, ps)
	}

	thresholdSig := getThresholdSignature(pss)

	return &typesCons.QuorumCertificate{
		Height:             block.BlockHeader.Height,
		Step:               1,
		Round:              1,
		Block:              block,
		ThresholdSignature: thresholdSig,
	}, nil
}

// generate partial signature for the validator
func generatePartialSignature(block *coreTypes.Block, node *shared.Node) (*typesCons.PartialSignature, error) {
	privKey, err := node.GetBus().GetConsensusModule().GetPrivateKey()
	if err != nil {
		return nil, err
	}
	return &typesCons.PartialSignature{
		Signature: getMessageSignature(block, privKey),
		Address:   privKey.PublicKey().Address().String(),
	}, nil
}

func getThresholdSignature(partialSigs []*typesCons.PartialSignature) *typesCons.ThresholdSignature {
	thresholdSig := new(typesCons.ThresholdSignature)
	thresholdSig.Signatures = make([]*typesCons.PartialSignature, len(partialSigs))
	copy(thresholdSig.Signatures, partialSigs)
	return thresholdSig
}

// Generates partial signature with given private key
// If there is an error signing the bytes, nil is returned instead.
func getMessageSignature(block *coreTypes.Block, privKey cryptoPocket.PrivateKey) []byte {
	bytesToSign, err := getSignableBytes(block)
	if err != nil {
		logger.Global.Warn().Err(err).Msgf("error getting bytes to sign")
		return nil
	}

	signature, err := privKey.Sign(bytesToSign)
	if err != nil {
		logger.Global.Warn().Err(err).Msgf("error signing message")
		return nil
	}

	return signature
}

// Signature only over subset of fields in HotstuffMessage
// For reference, see section 4.3 of the the hotstuff whitepaper, partial signatures are
// computed over `tsignr(hm.type, m.viewNumber , m.nodei)`. https://arxiv.org/pdf/1803.05069.pdf
func getSignableBytes(block *coreTypes.Block) ([]byte, error) {
	msgToSign := &typesCons.HotstuffMessage{
		Height: block.BlockHeader.Height,
		Step:   1,
		Round:  1,
		Block:  block,
	}
	return codec.GetCodec().Marshal(msgToSign)
}
