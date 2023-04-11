package utility

// IMPORTANT: The interface and implementation defined in this file are for illustrative purposes only
// and need to be revisited before any implementation commences.

import (
	"fmt"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
)

// TODO: When implementing please review if block height tolerance (+,-1) is included in the session protocol: pokt-network/pocket-core#1464 CC @Olshansk

type sessionHydrator struct {
	logger  modules.Logger
	session *coreTypes.Session
	readCtx modules.PersistenceReadContext
}

func (m *utilityModule) GetSession(appAddr string, height int64, relayChain coreTypes.RelayChain, geoZone string) (*coreTypes.Session, error) {
	persistenceModule := m.GetBus().GetPersistenceModule()
	readCtx, err := persistenceModule.NewReadContext(height)
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()

	session := &coreTypes.Session{
		Height:     height,
		RelayChain: relayChain,
		GeoZone:    geoZone,
	}

	sessionHydrator := &sessionHydrator{
		logger:  m.logger.With().Str("source", "sessionHydrator").Logger(),
		session: session,
		readCtx: readCtx,
	}

	if err := sessionHydrator.hydrateSessionId(); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionApplication(); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionServicers(); err != nil {
		return nil, err
	}

	if err := sessionHydrator.hydrateSessionFishermen(); err != nil {
		return nil, err
	}

	return sessionHydrator.session, nil
}

func getSessionHeight(readCtx modules.PersistenceReadContext, blockHeight int64) (int64, error) {
	return blockHeight, nil
}

// use the seed information to determine a SHA3Hash that is used to find the closest N actors based
// by comparing the sessionKey with the actors' public key
func (s *sessionHydrator) hydrateSessionId() error {
	s.readCtx.GetAppExists()
	// sessionIdBz := make([]byte, 8)
	// binary.LittleEndian.PutUint64(sessionHeightBz, uint64(s.height))

	// blockHashBz := []byte("get block hash bytes at s.sessionHeight from persistence module")

	// appPubKey, err := crypto.NewPublicKey(s.application.GetPublicKey())
	// if err != nil {
	// 	return nil, types.ErrNewPublicKeyFromBytes(err)
	// }

	// return concat(sessionHeightBz, blockHashBz, []byte(s.geoZone), []byte(s.relayChain), appPubKey.Bytes()), nil
	return nil
}

func (s *sessionHydrator) hydrateSessionApplication() error {
	// IMPORTANT: This function is for behaviour illustrative purposes only and implementation may differ.
	return nil
}

// uses the current 'world state' to determine the servicers in the session
// 1) get an ordered list of the public keys of servicers who are:
//   - actively staked
//   - staked within geo-zone (or closest geo-zones)
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(servicers, numberOfNodesPerSession)`
func (s *sessionHydrator) hydrateSessionServicers() error {
	// number of servicers per session at this height
	numServicers, err := s.readCtx.GetIntParam(types.ServicersPerSessionParamName, s.session.Height)
	if err != nil {
		return err
	}
	// s.session.Servicers = make([]*coreTypes.Actor, numServicers)

	// returns all the staked servicers at this session height
	servicers, err := s.readCtx.GetAllServicers(s.session.Height)
	if err != nil {
		return err
	}

	// OPTIMIZE: Update the persistence module to allow for querying for filtered servicers directly
	// Determine the servicers for this session
	candidateServicers := make([]*coreTypes.Actor, 0)
	for _, servicer := range servicers {
		// Sanity check the servicer is not paused or unstaking
		if servicer.PausedHeight == -1 || servicer.UnstakingHeight == -1 {
			return fmt.Errorf("selectSessionServicers should not have encountered a paused or unstaking servicer: %s", servicer.Address)
		}

		// TODO_IN_THIS_COMMIT: if servicer.GeoZone includes session.GeoZone

		// OPTIMIZE: If this was a map, we could have avoided the loop over chains
		var chain string
		for _, chain = range servicer.Chains {
			// TODO_IN_THIS_COMMIT: Change actor chains to use the enum
			if chain != string(s.session.RelayChain) {
				chain = ""
				continue
			}
		}
		if chain != "" {
			candidateServicers = append(candidateServicers, servicer)
		}
	}

	s.session.Servicers = s.pseudoRandomSelection(candidateServicers, numServicers)
	return nil
}

// uses the current 'world state' to determine the fishermen in the session
// 1) get an ordered list of the public keys of fishermen who are:
//   - actively staked
//   - staked within geo-zone  (or closest geo-zones)
//   - staked for relay-chain
//
// 2) calls `pseudoRandomSelection(fishermen, numberOfFishPerSession)`
func (s *sessionHydrator) hydrateSessionFishermen() error {
	// IMPORTANT: This function is for behaviour illustrative purposes only and implementation may differ.
	return nil
}

// 1) passed an ordered list of the public keys of actors and number of nodes
// 2) pseudo-insert the session `key` string into the list and find the first actor directly below
// 3) newKey = Hash( key + actor1PublicKey )
// 4) repeat steps 2 and 3 until all N actor are found
// FAQ:
// Q) why do we hash to find a newKey between every actor selection?
// A) pseudo-random selection only works if each iteration is re-randomized
//
//	or it would be subject to lexicographical proximity bias attacks
func (s *sessionHydrator) pseudoRandomSelection(candidates []*coreTypes.Actor, numTarget int) []*coreTypes.Actor {
	if numTarget < len(candidates) {
		s.logger.Warn().Msgf("pseudoRandomSelection: numTarget (%d) is less than the number of candidates (%d)", numTarget, len(candidates))
	}
	// TODO_IN_THIS_COMMIT: Actually implement this
	return candidates[:numTarget]
}

// func concat(b ...[]byte) (result []byte) {
// 	for _, bz := range b {
// 		result = append(result, bz...)
// 	}
// 	return result
// }
