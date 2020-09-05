// Package sqlite_db contains common types and functions for storing data in a sqlite database.
package sqlite_db

import (
	"errors"
	"github.com/keep94/appcommon/date_util"
	"github.com/keep94/appcommon/db"
	"github.com/keep94/gosqlite/sqlite"
	"time"
)

const (
	LastRowIdSQL = "select last_insert_rowid()"
)

var (
	AlreadyClosed = errors.New("sqlite_db: Already Closed")
	NoResult      = errors.New("sqlite_db: No result")
)

// Action represents some action against a sqlite database
type Action func(conn *sqlite.Conn) error

// Db wraps a sqlite database connection.
// With Db, multiple threads can safely share the same connection.
// Db also provides transactional behavior.
type Db struct {
	conn       connWrapper
	requestCh  chan Action
	responseCh chan error
	doneCh     chan struct{}
}

// New creates a new Db.
func New(conn *sqlite.Conn) *Db {
	return new(realConnWrapper{conn})
}

// Do performs action within a transaction. Do returns AlreadyClosed
// if Close was already called on this Db.
func (d *Db) Do(action Action) error {
	select {
	case <-d.doneCh:
		return AlreadyClosed
	case d.requestCh <- action:
		return <-d.responseCh
	}
	return nil
}

// Close closes the underlying connection.
func (d *Db) Close() error {
	return d.Do(nil)
}

func new(conn connWrapper) *Db {
	result := &Db{conn, make(chan Action), make(chan error), make(chan struct{})}
	go result.loop()
	return result
}

func (d *Db) loop() {
	action := <-d.requestCh
	for ; action != nil; action = <-d.requestCh {
		d.responseCh <- d.execute(action)
	}
	d.responseCh <- d.conn.Close()
	close(d.responseCh)
	close(d.doneCh)
}

func (d *Db) execute(action Action) error {
	err := d.conn.begin()
	if err != nil {
		return err
	}
	err = action(d.conn.delegate())
	if err != nil {
		d.conn.rollback()
		return err
	}
	err = d.conn.commit()
	if err != nil {
		d.conn.rollback()
		return err
	}
	return nil
}

// LastRowId returns the Id of the last inserted row in database.
func LastRowId(conn *sqlite.Conn) (id int64, err error) {
	stmt, err := conn.Prepare(LastRowIdSQL)
	if err != nil {
		return
	}
	defer stmt.Finalize()
	return LastRowIdFromStmt(stmt)
}

// LastRowIdFromStmt returns the Id of the last inserted row in database.
// stmt should be a prepared statement created from LastRowIdSQL.
func LastRowIdFromStmt(stmt *sqlite.Stmt) (id int64, err error) {
	if err = stmt.Exec(); err != nil {
		return
	}
	if !stmt.Next() {
		err = NoResult
		return
	}
	stmt.Scan(&id)
	return
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

func NewSqliteDoer(conn *sqlite.Conn) Doer {
	return simpleDoer{conn}
}

// Doer does an Action against a sqlite database
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
	return g.db.Do(func(conn *sqlite.Conn) error {
		return action(toTransaction(conn))
	})
}

func toTransaction(conn *sqlite.Conn) db.Transaction {
	return simpleDoer{conn}
}

type simpleDoer struct {
	conn *sqlite.Conn
}

func (s simpleDoer) Do(a Action) error {
	return a(s.conn)
}

type connWrapper interface {
	begin() error
	commit() error
	rollback() error
	delegate() *sqlite.Conn
	Close() error
}

type realConnWrapper struct {
	*sqlite.Conn
}

func (w realConnWrapper) begin() error {
	return w.Exec("begin")
}

func (w realConnWrapper) commit() error {
	return w.Exec("commit")
}

func (w realConnWrapper) rollback() error {
	return w.Exec("rollback")
}

func (w realConnWrapper) delegate() *sqlite.Conn {
	return w.Conn
}
