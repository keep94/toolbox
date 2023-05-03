// Package sqlite3_db contains common types and functions for storing data in a sqlite3 database.
package sqlite3_db

import (
	"database/sql"
	"sync"
	"time"

	"github.com/keep94/toolbox/date_util"
	"github.com/keep94/toolbox/db"
)

// Action represents some action against a sqlite3 database
type Action func(tx *sql.Tx) error

// Db wraps a sqlite3 database connection.
// With Db, multiple goroutines can safely share the same connection.
// Db also provides transactional behavior.
type Db struct {
	mu sync.Mutex
	db *sql.DB
}

// New creates a new Db.
func New(db *sql.DB) *Db {
	return &Db{db: db}
}

// Do performs action within a transaction.
func (d *Db) Do(action Action) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	err = action(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// Close closes the underlying sql.DB instance.
func (d *Db) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.db.Close()
}

// DateToString converts a date to YYYYmmdd
func DateToString(t time.Time) string {
	return t.Format(date_util.YMDFormat)
}

// StringToDate converts a string of form YYYYmmdd to a time object in UTC
// time zone.
func StringToDate(s string) (t time.Time, e error) {
	if s == "00010101" {
		return time.Time{}, nil
	}
	return time.Parse(date_util.YMDFormat, s)
}

func NewDoer(db *Db) db.Doer {
	return genericDoer{db}
}

func NewSqlite3Doer(tx *sql.Tx) Doer {
	return simpleDoer{tx}
}

// Doer does an Action against a sqlite3 database
type Doer interface {
	Do(Action) error
}

// If t is not nil, converts t to a Doer. Otherwise
// returns db as the Doer.
func ToDoer(db Doer, t db.Transaction) Doer {
	if t == nil {
		return db
	}
	return t.(Doer)
}

type genericDoer struct {
	db *Db
}

func (g genericDoer) Do(action db.Action) error {
	return g.db.Do(func(tx *sql.Tx) error {
		return action(toTransaction(tx))
	})
}

func toTransaction(tx *sql.Tx) db.Transaction {
	return simpleDoer{tx}
}

type simpleDoer struct {
	tx *sql.Tx
}

func (s simpleDoer) Do(a Action) error {
	return a(s.tx)
}
