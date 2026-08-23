package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Runner struct{ DB *sql.DB }

func (r Runner) Do(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	if r.DB == nil {
		return errors.New("database is nil")
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		if rb := tx.Rollback(); rb != nil {
			return fmt.Errorf("%w; rollback: %v", err, rb)
		}
		return err
	}
	return tx.Commit()
}
func Savepoint(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, "SAVEPOINT "+name)
	return err
}
func RollbackTo(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name)
	return err
}
func Release(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT "+name)
	return err
}
