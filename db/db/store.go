package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

// Store provides all functions to execute SQL queries
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new store
func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Store{
		Queries: New(db),
		db:      db,
	}, nil
}

// ExecTx executes a function within a database transaction
func (store *Store) ExecTxContext(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
