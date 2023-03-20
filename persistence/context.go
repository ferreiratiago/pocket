package persistence

// TECHDEBT: Look into whether the receivers of `PostgresContext` could/should be pointers?

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.PersistenceRWContext = &PostgresContext{}

// TECHDEBT: All the functions of `PostgresContext` should be organized in appropriate packages and use pointer receivers
type PostgresContext struct {
	logger *modules.Logger

	Height int64 // TECHDEBT: `Height` is only externalized for testing purposes. Replace with a `Debug` interface containing helpers

	conn *pgxpool.Conn
	tx   pgx.Tx

	stateHash string
	// TECHDEBT(#361): These three values are pointers to objects maintained by the PersistenceModule.
	//                 Need to simply access them via the bus.
	blockStore kvstore.KVStore
	txIndexer  indexer.TxIndexer
	stateTrees *stateTrees
}

func (p *PostgresContext) NewSavePoint(bytes []byte) error {
	p.logger.Info().Bool("TODO", true).Msg("NewSavePoint not implemented")
	return nil
}

// TECHDEBT(#327): Guarantee atomicity betweens `prepareBlock`, `insertBlock` and `storeBlock` for save points & rollbacks.
func (p *PostgresContext) RollbackToSavePoint(bytes []byte) error {
	p.logger.Info().Bool("TODO", true).Msg("RollbackToSavePoint not fully implemented")
	return p.tx.Rollback(context.TODO())
}

// IMPROVE(#361): Guarantee the integrity of the state
// Full details in the thread from the PR review: https://github.com/pokt-network/pocket/pull/285#discussion_r1018471719
func (p *PostgresContext) ComputeStateHash() (string, error) {
	stateHash, err := p.updateMerkleTrees()
	if err != nil {
		return "", err
	}
	p.stateHash = stateHash
	return p.stateHash, nil
}

// TECHDEBT(#327): Make sure these operations are atomic
func (p *PostgresContext) Commit(proposerAddr, quorumCert []byte) error {
	p.logger.Info().Int64("height", p.Height).Msg("About to commit block & context")

	// Create a persistence block proto
	block, err := p.prepareBlock(proposerAddr, quorumCert)
	if err != nil {
		return err
	}

	// Store block in the KV store
	if err := p.storeBlock(block); err != nil {
		return err
	}

	// Insert the block into the SQL DB
	if err := p.insertBlock(block); err != nil {
		return err
	}

	// Commit the SQL transaction
	ctx := context.TODO()
	if err := p.tx.Commit(ctx); err != nil {
		return err
	}

	p.conn.Release()

	return nil
}

func (p *PostgresContext) Release() error {
	p.logger.Info().Int64("height", p.Height).Msg("About to release context")
	ctx := context.TODO()
	if err := p.tx.Rollback(ctx); err != nil {
		return err
	}
	if err := p.resetContext(); err != nil {
		return err
	}
	return nil
}

func (p *PostgresContext) Close() error {
	if err := p.resetContext(); err != nil {
		return err
	}
	p.conn.Release()
	return nil
}

// INVESTIGATE(#361): Revisit if is used correctly in the context of the lifecycle of a persistenceContext and a utilityContext
func (p *PostgresContext) IndexTransaction(txResult modules.TxResult) error {
	return p.txIndexer.Index(txResult)
}

func (pg *PostgresContext) resetContext() (err error) {
	if pg == nil {
		pg.logger.Warn().Msg("postgres context is nil when trying to reset it")
		return nil
	}
	if pg.tx == nil {
		return nil
	}
	conn := pg.tx.Conn()
	if conn != nil && !conn.IsClosed() {
		if err := pg.Release(); err != nil {
			pg.logger.Error().Err(err).Bool("TODO", true).Msg("error releasing write context")
		}
	}
	pg.tx = nil
	return nil
}
