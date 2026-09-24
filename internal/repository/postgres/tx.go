package postgres

import (
	"context"
	"database/sql"

	"github.com/ran-jita/task-api/internal/repository"
)

type sqlTx struct {
	tx *sql.Tx
}

func (s *sqlTx) Commit() error   { return s.tx.Commit() }
func (s *sqlTx) Rollback() error { return s.tx.Rollback() }

type txManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) repository.TxManager {
	return &txManager{db: db}
}

func (m *txManager) Begin(ctx context.Context) (repository.Tx, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx: tx}, nil
}

func unwrap(tx repository.Tx) *sql.Tx {
	return tx.(*sqlTx).tx
}
