// Package sqlite3_rw reads and updates sqlite3 databases using consumers
// from the github.com/keep94/consume2 package.
package sqlite3_rw

import (
	"database/sql"
	"fmt"
	"hash/fnv"

	"github.com/keep94/consume2"
)

// RowForReading reads a single database row into its business object.
// RowForReading instances can optionally implement EtagSetter if
// its business object has an etag.
type RowForReading interface {

	// Ptrs returns the pointers to be passed to Scan to read the database row.
	Ptrs() []interface{}

	// Unmarshall updates this instance's business object with the values
	// stored in the pointers that Ptrs returned.
	Unmarshall() error
}

// RowsForReading is for reading multiple rows.
type RowsForReading[T any] interface {
	RowForReading

	// ValueRead returns the actual value of the business object just read
	// from the last row.
	ValueRead() T
}

// RowsForReadingEtagSetter handles both reading multiple rows and setting
// etags.
type RowsForReadingEtagSetter[T any] interface {
	RowsForReading[T]
	EtagSetter
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
	tx *sql.Tx,
	row RowForReading,
	noSuchRow error,
	sql string,
	params ...interface{}) error {
	dbrows, err := tx.Query(sql, params...)
	if err != nil {
		return err
	}
	defer dbrows.Close()
	return FirstOnly(row, dbrows, noSuchRow)
}

// FirstOnly reads one row from dbrows into row's business object. FirstOnly
// returns noSuchRow if dbrows has no rows.
func FirstOnly(
	row RowForReading,
	dbrows *sql.Rows,
	noSuchRow error) error {
	ptrs := row.Ptrs()
	rowRead := false
	if dbrows.Next() {
		if err := readRow(row, dbrows, ptrs, true); err != nil {
			return err
		}
		rowRead = true
	}
	if err := dbrows.Err(); err != nil {
		return err
	}
	if !rowRead {
		return noSuchRow
	}
	return nil
}

// ReadRows reads many rows from dbrows. For each row read, ReadRows adds
// row's business object to consumer. ReadRows does not set the etag in
// business objects read even if row implements EtagSetter.
func ReadRows[T any](
	row RowsForReading[T],
	dbrows *sql.Rows,
	consumer consume2.Consumer[T]) error {
	if err := readRows(row, dbrows, consumer, false); err != nil {
		return err
	}
	return dbrows.Err()
}

// ReadRowsWithEtag works like ReadRows except it does set the etag in
// business objects read.
func ReadRowsWithEtag[T any](
	row RowsForReadingEtagSetter[T],
	dbrows *sql.Rows,
	consumer consume2.Consumer[T]) error {
	if err := readRows[T](row, dbrows, consumer, true); err != nil {
		return err
	}
	return dbrows.Err()
}

func readRows[T any](
	row RowsForReading[T],
	dbrows *sql.Rows,
	consumer consume2.Consumer[T],
	setEtag bool) error {
	ptrs := row.Ptrs()
	for dbrows.Next() && consumer.CanConsume() {
		if err := readRow(row, dbrows, ptrs, setEtag); err != nil {
			return err
		}
		consumer.Consume(row.ValueRead())
	}
	return nil
}

// ReadMultiple executes sql and reads multiple rows. Each time a row
// is read, row's business object is added to consumer. params provides
// values for question mark (?) place holders in sql. ReadMultiple does
// not set the etag in business objects read even if row implements
// EtagSetter.
func ReadMultiple[T any](
	tx *sql.Tx,
	row RowsForReading[T],
	consumer consume2.Consumer[T],
	sql string,
	params ...interface{}) error {
	dbrows, err := tx.Query(sql, params...)
	if err != nil {
		return err
	}
	defer dbrows.Close()
	if err := readRows(row, dbrows, consumer, false); err != nil {
		return err
	}
	return dbrows.Err()
}

// ReadMultipleWithEtag works like ReadMultiple, but it also computes
// etags for fetched rows.
func ReadMultipleWithEtag[T any](
	tx *sql.Tx,
	row RowsForReadingEtagSetter[T],
	consumer consume2.Consumer[T],
	sql string,
	params ...interface{}) error {
	dbrows, err := tx.Query(sql, params...)
	if err != nil {
		return err
	}
	defer dbrows.Close()
	if err := readRows[T](row, dbrows, consumer, true); err != nil {
		return err
	}
	return dbrows.Err()
}

// AddRow adds row's business object as a new row in database.
// The row being added must have auto increment id field. AddRow stores the
// id of the new row at rowId.
func AddRow(
	tx *sql.Tx,
	row RowForWriting,
	rowId *int64,
	sql string) error {
	values, err := InsertValues(row)
	if err != nil {
		return err
	}
	result, err := tx.Exec(sql, values...)
	if err != nil {
		return err
	}
	*rowId, err = result.LastInsertId()
	return err
}

// UpdateRow updates a row's business object in the database.
func UpdateRow(
	tx *sql.Tx,
	row RowForWriting,
	sql string) error {
	values, err := UpdateValues(row)
	if err != nil {
		return err
	}
	_, err = tx.Exec(sql, values...)
	return err
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
	row RowForReading,
	dbrows *sql.Rows,
	ptrs []interface{},
	setEtag bool) error {
	if err := dbrows.Scan(ptrs...); err != nil {
		return err
	}
	if setEtag {
		etagSetter, isEtagSetter := row.(EtagSetter)
		if isEtagSetter {
			if err := doEtag(etagSetter); err != nil {
				return err
			}
		}
	}
	if err := row.Unmarshall(); err != nil {
		return err
	}
	return nil
}
