package runtime

import (
	"encoding/json"
	"os"

	"github.com/pokt-network/pocket/runtime/genesis"
)

// parseGenesis parses the genesis file in JSON format and returns a genesis.GenesisState
func parseGenesis(genesisJSONPath string) (g *genesis.GenesisState, err error) {
	data, err := os.ReadFile(genesisJSONPath)
	if err != nil {
		return
	}

	// general genesis file
	g = new(genesis.GenesisState)
	err = json.Unmarshal(data, &g)
	return
}
