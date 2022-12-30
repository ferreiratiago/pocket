package cli

import (
	"log"
	"os"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/p2p"
	debugABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/debug"
	debugCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/debug"
	"github.com/pokt-network/pocket/runtime"
	pocketCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
)

// TECHDEBT: Lowercase variables / constants that do not need to be exported.
const (
	PromptResetToGenesis         string = "ResetToGenesis"
	PromptPrintNodeState         string = "PrintNodeState"
	PromptTriggerNextView        string = "TriggerNextView"
	PromptTogglePacemakerMode    string = "TogglePacemakerMode"
	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore"

	defaultConfigPath  = "build/config/config1.json"
	defaultGenesisPath = "build/config/genesis.json"
)

var (
	// A P2P module is initialized in order to broadcast a message to the local network
	p2pMod modules.P2PModule

	// A consensus module is initialized in order to get a list of the validator network
	consensusMod modules.ConsensusModule
	modInitOnce  sync.Once

	items = []string{
		PromptResetToGenesis,
		PromptPrintNodeState,
		PromptTriggerNextView,
		PromptTogglePacemakerMode,
		PromptShowLatestBlockInStore,
	}

	validators []modules.Actor
)

func init() {
	rootCmd.AddCommand(NewDebugCommand())
}

func NewDebugCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug utility for rapid development",
		Args:  cobra.ExactArgs(0),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initDebug(remoteCLIURL)
		},
		RunE: runDebug,
	}
}

func runDebug(cmd *cobra.Command, args []string) (err error) {
	for {
		if selection, err := promptGetInput(); err == nil {
			handleSelect(selection)
		} else {
			return err
		}
	}
}

func promptGetInput() (string, error) {
	prompt := promptui.Select{
		Label: "Select an action",
		Items: items,
		Size:  len(items),
	}

	_, result, err := prompt.Run()

	if err == promptui.ErrInterrupt {
		os.Exit(0)
	}

	if err != nil {
		log.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}

func handleSelect(selection string) {
	switch selection {
	case PromptResetToGenesis:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptPrintNodeState:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTriggerNextView:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTogglePacemakerMode:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptShowLatestBlockInStore:
		m := &messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			Message: nil,
		}
		sendDebugMessage(m)
	default:
		log.Println("Selection not yet implemented...", selection)
	}
}

// Broadcast to the entire validator set
func broadcastDebugMessage(debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto, messaging.PocketTopic_DEBUG_TOPIC)

	for _, valAddr := range validators {
		addr, err := pocketCrypto.NewAddress(valAddr.GetAddress())
		if err != nil {
			log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
		}
		p2pMod.Send(addr, anyProto)
	}
}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(debugMsg *messaging.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	var validatorAddress []byte
	if len(validators) == 0 {
		log.Fatalf("[ERROR] No validators found")
	}

	validatorAddress, err = pocketCrypto.NewAddress(validators[0].GetAddress())
	if err != nil {
		log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
	}

	p2pMod.Send(validatorAddress, anyProto)
}

func initDebug(remoteCLIURL string) {
	modInitOnce.Do(func() {
		var err error
		runtimeMgr := runtime.NewManagerFromFiles(defaultConfigPath, defaultGenesisPath, runtime.WithRandomPK())

		// HACK: this is a temporary solution that guarantees backward compatibility while we implement peer discovery (#416).
		validators = runtimeMgr.GetGenesis().GetConsensusGenesisState().GetVals()

		debugAddressBookProvider := debugABP.NewDebugAddrBookProvider(
			runtimeMgr.GetConfig().GetP2PConfig(),
			debugABP.WithActorsByHeight(
				map[int64][]modules.Actor{
					debugABP.ALL_HEIGHTS: validators,
				},
			),
		)

		debugCurrentHeightProvider := debugCHP.NewDebugCurrentHeightProvider(0)

		p2pM, err := p2p.CreateWithProviders(runtimeMgr.GetBus(), debugAddressBookProvider, debugCurrentHeightProvider)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create p2p module: %v", err.Error())
		}
		p2pMod = p2pM.(modules.P2PModule)

		p2pMod.Start()
	})
}
