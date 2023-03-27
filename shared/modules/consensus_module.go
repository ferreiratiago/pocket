package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	ConsensusModuleName      = "consensus"
	PacemakerModuleName      = "pacemaker"
	LeaderElectionModuleName = "leader_election"
)

// NOTE: Consensus is the core of the replicated state machine and is driven by various asynchronous events.
// Consider adding a mutex lock to your implementation that is acquired at the beginning of each entrypoint/function implemented in this interface.
// Make sure that you are not locking again within the same call to avoid deadlocks (for example when the methods below call each other in your implementation).
type ConsensusModule interface {
	Module
	KeyholderModule

	ConsensusStateSync
	ConsensusPacemaker
	ConsensusDebugModule

	// Consensus Engine Handlers
	// TODO: Rename `HandleMessage` to a more specific name that is consistent with its business logic.
	HandleMessage(*anypb.Any) error
	// State Sync messages Handler
	HandleStateSyncMessage(*anypb.Any) error
	// FSM transition event handler
	HandleEvent(transitionMessageAny *anypb.Any) error

	// Consensus State Accessors
	CurrentHeight() uint64
	CurrentRound() uint64
	CurrentStep() uint64
}

// This interface represents functions exposed by the Consensus module for Pacemaker specific business logic.
// These functions are intended to only be called by the Pacemaker module.
// TODO(#428): This interface will be removed when the communication between the pacemaker and consensus module become asynchronous via the bus.
type ConsensusPacemaker interface {
	// Clearers
	ResetRound()
	ResetForNewHeight()
	ClearLeaderMessagesPool()
	// TODO: @deblasis - remove this and implement an event based approach
	ReleaseUtilityUnitOfWork() error

	// Setters
	SetHeight(uint64)
	SetRound(uint64)
	SetStep(uint8) // CONSIDERATION: Change to `typesCons.HotstuffStep; causes an import cycle.

	// Communicators
	BroadcastMessageToValidators(*anypb.Any) error

	// Leader helpers
	IsLeader() bool
	IsLeaderSet() bool
	NewLeader(*anypb.Any) error // CONSIDERATION: Consider changing input to typesCons.HotstuffMessage. This requires to do refactoring.

	// Getters
	IsPrepareQCNil() bool
	GetPrepareQC() (*anypb.Any, error)
	GetNodeId() uint64
}

// This interface represents functions exposed by the Consensus module for StateSync specific business logic.
// These functions are intended to only be called by the StateSync module.
// INVESTIGATE: This interface enable a fast implementation of state sync but look into a way of removing it in the future
type ConsensusStateSync interface {
	GetNodeIdFromNodeAddress(string) (uint64, error)
	GetNodeAddress() string

	// IsSynched compares the persisted state with the aggregated state of the network. If the persisted state is behind the network state, i.e. that node is not synched, it will return false.
	IsSynched() (bool, error)
}

// This interface represents functions exposed by the Consensus module for mainly used for testing.
// This interface is not intended to be used by any other module than testing in Consensus module.
type ConsensusDebugModule interface {
	HandleDebugMessage(*messaging.DebugMessage) error

	SetHeight(uint64)
	SetRound(uint64)
	// REFACTOR: This should accept typesCons.HotstuffStep.
	SetStep(uint8)
	SetBlock(*types.Block)

	SetUtilityUnitOfWork(UtilityUnitOfWork)

	SetAggregatedStateSyncMetadata(minHeight, maxHeight uint64, peerAddress string)
	GetAggregatedStateSyncMetadataMaxHeight() (minHeight uint64)

	// REFACTOR: This should accept typesCons.HotstuffStep and return typesCons.NodeId.
	GetLeaderElectionResult(height, round uint64, step uint8) (leaderId uint64)
}
