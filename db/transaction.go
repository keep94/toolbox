// Package db provides general datastore support.
package db

// Transaction represents a database transaction. When a nil Transaction is
// passed to a datastore operation, it means run the operation in its own
// transaction.
type Transaction interface{}

// Action represents a sequence of operations to be done in a single transaction.
// t is that single transaction. The action must pass t as the transaction to
// each datastore operation.
type Action func(t Transaction) error

// Doer performs an action within a single transaction. Each implementation
// specific database package underneath this package has a NewDoer method 
// that creates an instance of this interface for that implementation.
type Doer interface {
  Do(action Action) error
}
