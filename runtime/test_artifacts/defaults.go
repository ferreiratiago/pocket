package test_artifacts

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/utils"
)

var (
	DefaultChains              = []string{"0001"}
	DefaultServiceURL          = ""
	DefaultStakeAmount         = big.NewInt(1000000000000)
	DefaultStakeAmountString   = utils.BigIntToString(DefaultStakeAmount)
	DefaultAccountAmount       = big.NewInt(100000000000000)
	DefaultAccountAmountString = utils.BigIntToString(DefaultAccountAmount)
	DefaultPauseHeight         = int64(-1) // pauseHeight=-1 implies not paused
	DefaultUnstakingHeight     = int64(-1) // unstakingHeight=-1 implies not unstaking
	DefaultChainID             = "testnet"
	ServiceURLFormat           = "node%d.consensus:42069"
	DefaultMaxBlockBytes       = uint64(4000000)
)
