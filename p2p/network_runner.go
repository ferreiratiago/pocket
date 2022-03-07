package p2p

import "github.com/pokt-network/pocket/p2p/types"

func (m *p2pModule) Sink() chan<- types.Work {
	return m.sink
}

func (m *p2pModule) Done() <-chan uint {
	return m.done
}
