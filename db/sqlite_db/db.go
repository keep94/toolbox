// Package sqlite_db contains common types and functions for storing data in a sqlite database.
package sqlite_db

import (
  "code.google.com/p/gosqlite/sqlite"
  "errors"
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "hash/fnv"
  "time"
)

const (
  LastRowIdSQL = "select last_insert_rowid()"
)

var (
  AlreadyClosed = errors.New("sqlite_db: Already Closed")
  NoResult = errors.New("sqlite_db: No result")
)

// Action represents some action against a sqlite database
type Action func(conn *sqlite.Conn) error

// RowForReading represents a table row with ID column for reading.
type RowForReading interface {
  functional.Tuple
  // Pair pairs this value with a business object. ptr points to the business
  // object.
  Pair(ptr interface{})
  // Unmarshall populates associated business object.
  Unmarshall() error
}

// RowForWriting represents a table row with ID column that is being written.
type RowForWriting interface {
  // Values returns columns in row with Id column last.
  Values() []interface{}
  // Pair pairs this value with a business object. ptr points to the business
  // object.
  Pair(ptr interface{})
  // Marshall populates columns from associated business object.
  Marshall() error
}
  
// SimpleRow provides empty Marshall / Unmarshall for implementations of
// RowForReading and RowForWriting
type SimpleRow struct {
}

func (s SimpleRow) Marshall() error {
  return nil
}

func (s SimpleRow) Unmarshall() error {
  return nil
}

// Db wraps a sqlite database connection.
// With Db, multiple threads can safely share the same connection.
// Db also provides transactional behavior.
type Db struct {
  conn connWrapper
  requestCh chan Action
  responseCh chan error
  doneCh chan struct{}
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

// InsertValues returns the values needed to insert a row in a table.
// row is the table row; ptr points to the business object to insert.
func InsertValues(row RowForWriting, ptr interface{}) (
    values []interface{}, err error) {
  values, err = UpdateValues(row, ptr)
  values = values[:len(values) - 1]
  return
}

// UpdateValues returns the values needed to update a row in a table.
// row is the table row; ptr points to the business object to update.
func UpdateValues(row RowForWriting, ptr interface{}) (
    values []interface{}, err error) {
  row.Pair(ptr)
  err = row.Marshall()
  values = row.Values()
  return
}

// ReadRows returns table rows as a Stream of business objects
// or of db.Etagger instances (for collecting the etags of the
// business objects) depending on what the caller passes to
// the Next method of the returned stream.
// row must also implement RowForWriting if caller treats returned stream
// as a stream of db.Etagger instances.
// Caller must explicitly call Finalize on stmt when finished
// with returned Stream. Calling Close on returned Stream does
// nothing.
func ReadRows(row RowForReading, stmt *sqlite.Stmt) functional.Stream {
  stream := functional.ReadRows(stmt)
  return &rowStream{Stream: stream, row: row}
}

// ReadSingle executes sql and reads a single row into the business object 
// or db.Etagger instance at ptr. For the latter case, row must also
// implement RowForWriting
func ReadSingle(
    conn *sqlite.Conn,
    row RowForReading,
    noSuchRow error,
    ptr interface{},
    sql string,
    params ...interface{}) error {
  return ReadMultiple(
      conn,
      row,
      functional.ConsumerFunc(func(s functional.Stream) error {
        return consume.FirstOnly(s, noSuchRow, ptr)
      }),
      sql,
      params...)
}

// ReadMultiple executes sql and reads multiple rows.
// consumer may consume either business objects or db.Etagger objects.
// For the latter case, row must also implement RowForWriting
func ReadMultiple(
    conn *sqlite.Conn,
    row RowForReading,
    consumer functional.Consumer,
    sql string,
    params ...interface{}) error {
  stmt, err := conn.Prepare(sql)
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  if err = stmt.Exec(params...); err != nil {
    return err
  }
  return consumer.Consume(ReadRows(row, stmt))
}

// AddRow adds a new row. row being added must have auto increment id field.
// ptr points to the business object being added.
func AddRow(
    conn *sqlite.Conn,
    row RowForWriting,
    ptr interface{},
    rowId *int64,
    sql string) error {
  values, err := InsertValues(row, ptr)
  if err != nil {
    return err
  }
  if err = conn.Exec(sql, values...); err != nil {
    return err
  }
  *rowId, err = LastRowId(conn)
  return err
}

// UpdateRow updates a row. ptr points to the business object being updated.
// sql must be of form "update table_name set ... where id = ?"
func UpdateRow(
    conn *sqlite.Conn,
    row RowForWriting,
    ptr interface{},
    sql string) error {
  values, err := UpdateValues(row, ptr)
  if err != nil {
    return err
  }
  return conn.Exec(sql, values...)
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

type rowStream struct {
  functional.Stream
  row RowForReading
}

func (s *rowStream) Next(ptr interface{}) error {
  etagger, isEtagger := ptr.(db.Etagger)
  if isEtagger {
    s.row.Pair(etagger.GetPtr())
  } else {
    s.row.Pair(ptr)
  }
  err := s.Stream.Next(s.row)
  if err != nil {
    return err
  }
  if isEtagger {
    writeRow := s.row.(RowForWriting)
    etag, err := computeEtag(writeRow.Values())
    if err != nil {
      return err
    }
    etagger.SetEtag(etag)
  }
  return s.row.Unmarshall()
}

func computeEtag(values interface{}) (uint64, error) {
  h := fnv.New64a()
  s := fmt.Sprintf("%v", values)
  _, err := h.Write(([]byte)(s))
  if err != nil {
    return 0, err
  }
  return h.Sum64(), nil
}
