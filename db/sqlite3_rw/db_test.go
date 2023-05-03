package sqlite3_rw_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/keep94/consume2"
	"github.com/keep94/toolbox/db/sqlite3_db"
	"github.com/keep94/toolbox/db/sqlite3_rw"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestSqlError(t *testing.T) {
	assert := assert.New(t)
	rawdb, _ := sql.Open("sqlite3", ":memory:")
	defer rawdb.Close()
	db := sqlite3_db.New(rawdb)
	var records []Record
	assert.Error(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultiple[Record](
			tx,
			(&rawRecordWithEtag{}).init(&Record{}),
			consume2.AppendTo(&records),
			"select id, name, phone from records where name = ? order by id asc",
			"a",
		)
	}))
}

func TestDatabase(t *testing.T) {
	assert := assert.New(t)
	rawdb, _ := sql.Open("sqlite3", ":memory:")
	defer rawdb.Close()
	db := sqlite3_db.New(rawdb)
	db.Do(createTable)
	rec1 := Record{Name: "a", Phone: "1"}
	rec2 := Record{Name: "b", Phone: "2"}
	rec3 := Record{Name: "a", Phone: "3"}
	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx,
			(&rawRecord{}).init(&rec1),
			&rec1.Id,
			"insert into records (name, phone) values (?, ?)",
		)
	}))
	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx,
			(&rawRecord{}).init(&rec2),
			&rec2.Id,
			"insert into records (name, phone) values (?, ?)",
		)
	}))
	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx,
			(&rawRecord{}).init(&rec3),
			&rec3.Id,
			"insert into records (name, phone) values (?, ?)",
		)
	}))
	assert.Equal(int64(1), rec1.Id)
	assert.Equal(int64(2), rec2.Id)
	assert.Equal(int64(3), rec3.Id)

	var records []Record

	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultiple[Record](
			tx,
			(&rawRecordWithEtag{}).init(&Record{}),
			consume2.AppendTo(&records),
			"select id, name, phone from records where name = ? order by id asc",
			"a",
		)
	}))

	assert.Len(records, 2)
	assert.Equal(int64(1), records[0].Id)
	assert.Equal("a", records[0].Name)
	assert.Equal("1", records[0].Phone)
	assert.Equal(uint64(0), records[0].Etag)
	assert.Equal(int64(3), records[1].Id)
	assert.Equal("a", records[1].Name)
	assert.Equal("3", records[1].Phone)
	assert.Equal(uint64(0), records[1].Etag)

	records = records[:0]
	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultipleWithEtag[Record](
			tx,
			(&rawRecordWithEtag{}).init(&Record{}),
			consume2.AppendTo(&records),
			"select id, name, phone from records where name = ? order by id asc",
			"a",
		)
	}))

	assert.Len(records, 2)
	assert.Equal(int64(1), records[0].Id)
	assert.Equal("a", records[0].Name)
	assert.Equal("1", records[0].Phone)
	assert.NotEqual(uint64(0), records[0].Etag)
	assert.Equal(int64(3), records[1].Id)
	assert.Equal("a", records[1].Name)
	assert.Equal("3", records[1].Phone)
	assert.NotEqual(uint64(0), records[1].Etag)

	noSuchId := errors.New("No such id")

	var fourthRecord Record

	assert.Equal(
		db.Do(func(tx *sql.Tx) error {
			return sqlite3_rw.ReadSingle(
				tx,
				(&rawRecordWithEtag{}).init(&fourthRecord),
				noSuchId,
				"select id, name, phone from records where id = ?",
				4,
			)
		}),
		noSuchId)

	var secondRecord Record

	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawRecordWithEtag{}).init(&secondRecord),
			noSuchId,
			"select id, name, phone from records where id = ?",
			2,
		)
	}))

	secondEtag := secondRecord.Etag

	secondRecord = Record{}

	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawRecordWithEtag{}).init(&secondRecord),
			noSuchId,
			"select id, name, phone from records where id = ?",
			2,
		)
	}))

	assert.Equal(int64(2), secondRecord.Id)
	assert.Equal("b", secondRecord.Name)
	assert.Equal("2", secondRecord.Phone)
	assert.Equal(secondEtag, secondRecord.Etag)

	secondRecord.Phone = "1234"

	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.UpdateRow(
			tx,
			(&rawRecord{}).init(&secondRecord),
			"update records set name = ?, phone = ? where id = ?",
		)
	}))

	secondRecord = Record{}

	assert.Nil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawRecordWithEtag{}).init(&secondRecord),
			noSuchId,
			"select id, name, phone from records where id = ?",
			2,
		)
	}))

	assert.Equal(int64(2), secondRecord.Id)
	assert.Equal("b", secondRecord.Name)
	assert.Equal("1234", secondRecord.Phone)
	assert.NotEqual(secondEtag, secondRecord.Etag)

	assert.NotNil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultiple[Record](
			tx,
			(&errorRecord{}).init(&Record{}),
			consume2.AppendTo(&records),
			"select id, name, phone from records",
		)
	}))

	assert.NotNil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx,
			(&errorRecord{}).init(&secondRecord),
			&secondRecord.Id,
			"insert into records (name, phone) values (?, ?)",
		)
	}))

	assert.NotNil(db.Do(func(tx *sql.Tx) error {
		return sqlite3_rw.UpdateRow(
			tx,
			(&errorRecord{}).init(&secondRecord),
			"update records set name = ?, phone = ? where id = ?",
		)
	}))
}

func createTable(tx *sql.Tx) error {
	_, err := tx.Exec("create table if not exists records (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, phone TEXT)")
	return err
}

type Record struct {
	Id    int64
	Name  string
	Phone string
	Etag  uint64
}

type rawRecord struct {
	sqlite3_rw.SimpleRow
	*Record
}

func (r *rawRecord) init(bo *Record) *rawRecord {
	r.Record = bo
	return r
}

func (r *rawRecord) Ptrs() []interface{} {
	return []interface{}{&r.Id, &r.Name, &r.Phone}
}

func (r *rawRecord) ValueRead() Record {
	return *r.Record
}

func (r *rawRecord) Values() []interface{} {
	return []interface{}{r.Name, r.Phone, r.Id}
}

type rawRecordWithEtag struct {
	rawRecord
}

func (r *rawRecordWithEtag) init(bo *Record) *rawRecordWithEtag {
	r.rawRecord.init(bo)
	return r
}

func (r *rawRecordWithEtag) SetEtag(etag uint64) {
	r.Etag = etag
}

type errorRecord struct {
	*Record
}

func (e *errorRecord) init(bo *Record) *errorRecord {
	e.Record = bo
	return e
}

func (e *errorRecord) Ptrs() []interface{} {
	return []interface{}{&e.Id, &e.Name, &e.Phone}
}

func (e *errorRecord) Values() []interface{} {
	return []interface{}{e.Name, e.Phone, e.Id}
}

func (e *errorRecord) ValueRead() Record {
	return *e.Record
}

func (e *errorRecord) Marshall() error {
	return errors.New("An error happened marshalling")
}

func (e *errorRecord) Unmarshall() error {
	return errors.New("An error happened unmarshalling")
}
