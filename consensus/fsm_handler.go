package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

// HandleEvent handles FSM state transition events.
func (m *consensusModule) HandleEvent(transitionMessageAny *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	// TODO (#571): update with logger helper function
	m.logger.Info().Msgf("Received a state transition message: %s", transitionMessageAny)

	switch transitionMessageAny.MessageName() {
	case messaging.StateMachineTransitionEventType:
		msg, err := codec.GetCodec().FromAny(transitionMessageAny)
		if err != nil {
			return err
		}

		stateTransitionMessage, ok := msg.(*messaging.StateMachineTransitionEvent)
		if !ok {
			return fmt.Errorf("failed to cast message to StateSyncMessage")
		}

		return m.handleStateTransitionEvent(stateTransitionMessage)
	default:
		return typesCons.ErrUnknownStateSyncMessageType(transitionMessageAny.MessageName())
	}
}

func (m *consensusModule) handleStateTransitionEvent(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Msgf("Begin handling StateMachineTransitionEvent: %s", msg)

	// TODO (#571): update with logger helper function
	fsm_state := msg.NewState
	m.logger.Debug().Fields(map[string]any{
		"event":          msg.Event,
		"previous_state": msg.PreviousState,
		"new_state":      fsm_state,
	}).Msg("Received state machine transition msg")

	switch coreTypes.StateMachineState(fsm_state) {
	case coreTypes.StateMachineState_P2P_Bootstrapped:
		return m.HandleBootstrapped(msg)

	case coreTypes.StateMachineState_Consensus_Unsynched:
		return m.HandleUnsynched(msg)

	case coreTypes.StateMachineState_Consensus_SyncMode:
		return m.HandleSyncMode(msg)

	case coreTypes.StateMachineState_Consensus_Synched:
		return m.HandleSynched(msg)

	case coreTypes.StateMachineState_Consensus_Pacemaker:
		return m.HandlePacemaker(msg)

	default:
		m.logger.Warn().Msgf("Consensus module not handling this event: %s", msg.Event)

	}

	return nil
}

// HandleBootstrapped handles FSM event P2P_IsBootstrapped, and P2P_Bootstrapped is the destination state.
// Bootrstapped mode is when the node (validator or non-validator) is first coming online.
// This is a transition mode from node bootstrapping to a node being out-of-sync.
func (m *consensusModule) HandleBootstrapped(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM is in bootstrapped state, so it is out of sync, and transitions to unsynched mode")
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsUnsynched)
}

// HandleUnsynched handles FSM event Consensus_IsUnsynched, and Unsynched is the destination state.
// In Unsynched mode node (validator or non-validator) is out of sync with the rest of the network.
// This mode is a transition mode from the node being up-to-date (i.e. Pacemaker mode, Synched mode) with the latest network height to being out-of-sync.
// As soon as node transitions to this mode, it will transition to the sync mode.
func (m *consensusModule) HandleUnsynched(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM is in Unsyched state, as node is out of sync sending syncmode event to start syncing")

	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncing)
}

// HandleSyncMode handles FSM event Consensus_IsSyncing, and SyncMode is the destination state.
// In Sync mode node (validator or non-validator) starts syncing with the rest of the network.
func (m *consensusModule) HandleSyncMode(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM is in Sync Mode, start syncing...")

	//return m.stateSync.StartSynching()

	return m.stateSync.TriggerSync()
}

// HandleSynched handles FSM event IsSynchedNonValidator for Non-Validators, and Synched is the destination state.
// Currently, FSM never transition to this state and a non-validator node always stays in syncmode.
// CONSIDER: when a non-validator sync is implemented, maybe there is a case that requires transitioning to this state.
// TODO: Add check that this never happens when IsValidator() is false, i.e. node is not validator.
func (m *consensusModule) HandleSynched(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM of non-validator node is in Synched mode")
	return nil
}

// HandlePacemaker handles FSM event IsSynchedValidator, and Pacemaker is the destination state.
// Execution of this state means the validator node is synched.
func (m *consensusModule) HandlePacemaker(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM of validator node is synched and in Pacemaker mode. It will stay in this mode until it receives a new block proposal that has a higher height than the current block height")
	// validator receives a new block proposal, and it understands that it doesn't have block and it transitions to unsycnhed state
	// transitioning out of this state happens when a new block proposal is received by the hotstuff_replica
	return nil
}