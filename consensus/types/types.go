package types

// TODO: Split this file into multiple types files.
import (
	"sort"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

type NodeId uint64

type ValAddrToIdMap map[string]NodeId // Mapping from hex encoded address to an integer node id.
type IdToValAddrMap map[NodeId]string // Mapping from node id to a hex encoded string address.
type ValidatorMap map[string]*coreTypes.Actor

type ConsensusNodeState struct {
	NodeId NodeId
	Height uint64
	Round  uint8
	Step   uint8

	LeaderId NodeId
	IsLeader bool
}

func GetValAddrToIdMap(validators []modules.Actor) (ValAddrToIdMap, IdToValAddrMap) {
	valAddresses := make([]string, 0, len(validators))
	for _, val := range validators {
		valAddresses = append(valAddresses, val.GetAddress())
	}
	sort.Strings(valAddresses)

	valToIdMap := make(ValAddrToIdMap, len(valAddresses))
	idToValMap := make(IdToValAddrMap, len(valAddresses))
	for i, addr := range valAddresses {
		nodeId := NodeId(i + 1)
		valToIdMap[addr] = nodeId
		idToValMap[nodeId] = addr
	}

	return valToIdMap, idToValMap
}

func ActorListToValidatorMap(actors []*coreTypes.Actor) (m ValidatorMap) {
	m = make(ValidatorMap, len(actors))
	for _, a := range actors {
		m[a.GetAddress()] = a
	}
	return
}
