// Package sqlite_rw reads and updates sqlite databases using consumers
// from the github.com/keep94/goconsume package.
package sqlite_rw

import (
	"fmt"
	"hash/fnv"

	"github.com/keep94/appcommon/db/sqlite_db"
	"github.com/keep94/goconsume"
	"github.com/keep94/gosqlite/sqlite"
)

// RowForReading reads a database row into its business object.
// RowForReading instances can optionally implement EtagSetter if
// its business object has an etag.
type RowForReading interface {

	// ValuePtr returns the pointer to this instance's business object.
	ValuePtr() interface{}

	// Ptrs returns the pointers to be passed to Scan to read the database row.
	Ptrs() []interface{}

	// Unmarshall updates this instance's business object with the values
	// stored in the pointers that Ptrs returned.
	Unmarshall() error
}

// EtagSetter sets the etag on its business objecct
type EtagSetter interface {

	// Values returns column values from database with Id column last
	Values() []interface{}

	// SetEtag sets the etag on this instance's business object
	SetEtag(etag uint64)
}

// RowForWriting writes its business object to a database row.
type RowForWriting interface {

	// Values returns the column values for the database with Id column last.
	Values() []interface{}

	// Marshall updates the values that Values() returns using this instance's
	// business object
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

// ReadSingle executes sql and reads a single row into row's business object.
// ReadSingle returns noSuchRow if no rows were found. params provides the
// values for the question mark (?) place holders in sql.
func ReadSingle(
	conn *sqlite.Conn,
	row RowForReading,
	noSuchRow error,
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
	return FirstOnly(row, stmt, noSuchRow)
}

// FirstOnly reads one row from stmt into row's business object. FirstOnly
// returns noSuchRow if stmt has no rows.
func FirstOnly(
	row RowForReading,
	stmt *sqlite.Stmt,
	noSuchRow error) error {
	ptrs := row.Ptrs()
	if stmt.Next() {
		if err := readRow(row, stmt, ptrs); err != nil {
			return err
		}
		return nil
	}
	return noSuchRow
}

// ReadRows reads many rows from stmt. For each row read, ReadRows adds
// row's business object to consumer.
func ReadRows(
	row RowForReading,
	stmt *sqlite.Stmt,
	consumer goconsume.Consumer) error {
	ptrs := row.Ptrs()
	for stmt.Next() && consumer.CanConsume() {
		if err := readRow(row, stmt, ptrs); err != nil {
			return err
		}
		consumer.Consume(row.ValuePtr())
	}
	return nil
}

// ReadMultiple executes sql and reads multiple rows. Each time a row
// is read, row's business object is added to consumer. params provides
// values for question mark (?) place holders in sql.
func ReadMultiple(
	conn *sqlite.Conn,
	row RowForReading,
	consumer goconsume.Consumer,
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
	return ReadRows(row, stmt, consumer)
}

// AddRow adds row's business object as a new row in database.
// The row being added must have auto increment id field. AddRow stores the
// id of the new row at rowId.
func AddRow(
	conn *sqlite.Conn,
	row RowForWriting,
	rowId *int64,
	sql string) error {
	values, err := InsertValues(row)
	if err != nil {
		return err
	}
	if err = conn.Exec(sql, values...); err != nil {
		return err
	}
	*rowId, err = sqlite_db.LastRowId(conn)
	return err
}

// UpdateRow updates a row's business object in the database.
func UpdateRow(
	conn *sqlite.Conn,
	row RowForWriting,
	sql string) error {
	values, err := UpdateValues(row)
	if err != nil {
		return err
	}
	return conn.Exec(sql, values...)
}

// UpdateValues returns the values of the SQL columns to update row
func UpdateValues(row RowForWriting) (
	values []interface{}, err error) {
	if err = row.Marshall(); err != nil {
		return
	}
	return row.Values(), nil
}

// InsertValues returns the values of the SQL columns to add a new row
func InsertValues(row RowForWriting) (
	values []interface{}, err error) {
	var valuesForUpdate []interface{}
	if valuesForUpdate, err = UpdateValues(row); err != nil {
		return
	}
	return valuesForUpdate[:len(valuesForUpdate)-1], nil
}

func doEtag(row EtagSetter) error {
	etag, err := computeEtag(row.Values())
	if err != nil {
		return err
	}
	row.SetEtag(etag)
	return nil
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

func readRow(
	row RowForReading, stmt *sqlite.Stmt, ptrs []interface{}) error {
	if err := stmt.Scan(ptrs...); err != nil {
		return err
	}
	etagSetter, isEtagSetter := row.(EtagSetter)
	if isEtagSetter {
		if err := doEtag(etagSetter); err != nil {
			return err
		}
	}
	if err := row.Unmarshall(); err != nil {
		return err
	}
	return nil
}
