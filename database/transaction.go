package database

import (
	"context"
	"fmt"
)

// TxFn is a function that takes in a transaction, and returns a result and an
// error.
type TxFn[Result any] func(ctx context.Context, tx Tx) (Result, error)

// Transaction executes a TxFn within a transaction.
// It recovers from panics, rolls back on errors, and commits if no error
// occurs.
//
// Parameters:
//   - ctx: The context for the transaction.
//   - tx: The transaction to use.
//   - txFn: The function to execute in a transaction.
//
// Returns:
//   - Result: The result of the transactional function.
//   - error: An error if the transaction fails.
func Transaction[Result any](
	ctx context.Context, tx Tx, txFn TxFn[Result],
) (result Result, txErr error) {
	defer func() {
		// Recover from panics.
		var recovered any
		panicOccurred := false
		if recovered = recover(); recovered != nil {
			panicOccurred = true
			txErr = fmt.Errorf("Transaction TxFn panicked: %v", recovered)
		}
		// Rollback or commit the transaction.
		if err := finalizeTransaction(tx, txErr); err != nil {
			txErr = err
			var zero Result
			result = zero
		}
		// Propagate the panic if there was one.
		if panicOccurred {
			panic(recovered)
		}
	}()
	return txFn(ctx, tx)
}

// finalizeTransaction commits or rollbacks a transaction.
func finalizeTransaction(tx Tx, txErr error) error {
	if txErr != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("finalizeTransaction rollback error: %w", err)
		}
		return nil
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("finalizeTransaction commit error: %w", err)
	}
	return nil
}
