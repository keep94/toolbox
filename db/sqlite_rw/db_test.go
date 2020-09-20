package sqlite_rw_test

import (
	"errors"
	"github.com/keep94/goconsume"
	"github.com/keep94/gosqlite/sqlite"
	"github.com/keep94/toolbox/db/sqlite_rw"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	beginFailed  = errors.New("sqlite_db: begin failed.")
	commitFailed = errors.New("sqlite_db: commit failed.")
)

func TestDatabase(t *testing.T) {
	assert := assert.New(t)
	conn, _ := sqlite.Open(":memory:")
	createTable(conn)
	rec1 := Record{Name: "a", Phone: "1"}
	rec2 := Record{Name: "b", Phone: "2"}
	rec3 := Record{Name: "a", Phone: "3"}
	assert.Nil(sqlite_rw.AddRow(
		conn,
		(&rawRecord{}).init(&rec1),
		&rec1.Id,
		"insert into records (name, phone) values (?, ?)",
	))
	assert.Nil(sqlite_rw.AddRow(
		conn,
		(&rawRecord{}).init(&rec2),
		&rec2.Id,
		"insert into records (name, phone) values (?, ?)",
	))
	assert.Nil(sqlite_rw.AddRow(
		conn,
		(&rawRecord{}).init(&rec3),
		&rec3.Id,
		"insert into records (name, phone) values (?, ?)",
	))
	assert.Equal(int64(1), rec1.Id)
	assert.Equal(int64(2), rec2.Id)
	assert.Equal(int64(3), rec3.Id)

	var records []Record

	assert.Nil(sqlite_rw.ReadMultiple(
		conn,
		(&rawRecord{}).init(&Record{}),
		goconsume.AppendTo(&records),
		"select id, name, phone from records where name = ? order by id asc", "a"))

	assert.Equal(int64(1), records[0].Id)
	assert.Equal("a", records[0].Name)
	assert.Equal("1", records[0].Phone)
	assert.Equal(uint64(0), records[0].Etag)
	assert.Equal(int64(3), records[1].Id)
	assert.Equal("a", records[1].Name)
	assert.Equal("3", records[1].Phone)
	assert.Equal(uint64(0), records[1].Etag)

	noSuchId := errors.New("No such id")

	var fourthRecord Record

	assert.Equal(sqlite_rw.ReadSingle(
		conn,
		(&rawRecordWithEtag{}).init(&fourthRecord),
		noSuchId,
		"select id, name, phone from records where id = ?", 4), noSuchId)

	var secondRecord Record

	assert.Nil(sqlite_rw.ReadSingle(
		conn,
		(&rawRecordWithEtag{}).init(&secondRecord),
		noSuchId,
		"select id, name, phone from records where id = ?", 2))

	secondEtag := secondRecord.Etag

	secondRecord = Record{}

	assert.Nil(sqlite_rw.ReadSingle(
		conn,
		(&rawRecordWithEtag{}).init(&secondRecord),
		noSuchId,
		"select id, name, phone from records where id = ?", 2))

	assert.Equal(int64(2), secondRecord.Id)
	assert.Equal("b", secondRecord.Name)
	assert.Equal("2", secondRecord.Phone)
	assert.Equal(secondEtag, secondRecord.Etag)

	secondRecord.Phone = "1234"

	assert.Nil(sqlite_rw.UpdateRow(
		conn,
		(&rawRecord{}).init(&secondRecord),
		"update records set name = ?, phone = ? where id = ?"))

	secondRecord = Record{}

	assert.Nil(sqlite_rw.ReadSingle(
		conn,
		(&rawRecordWithEtag{}).init(&secondRecord),
		noSuchId,
		"select id, name, phone from records where id = ?", 2))

	assert.Equal(int64(2), secondRecord.Id)
	assert.Equal("b", secondRecord.Name)
	assert.Equal("1234", secondRecord.Phone)
	assert.NotEqual(secondEtag, secondRecord.Etag)

	assert.NotNil(sqlite_rw.ReadMultiple(
		conn,
		(&errorRecord{}).init(&Record{}),
		goconsume.AppendTo(&records),
		"select id, name, phone from records"))

	assert.NotNil(sqlite_rw.AddRow(
		conn,
		(&errorRecord{}).init(&secondRecord),
		&secondRecord.Id,
		"insert into records (name, phone) values (?, ?)"))

	assert.NotNil(sqlite_rw.UpdateRow(
		conn,
		(&errorRecord{}).init(&secondRecord),
		"update records set name = ?, phone = ? where id = ?"))
}

func createTable(conn *sqlite.Conn) {
	conn.Exec("create table if not exists records (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, phone TEXT)")
}

type Record struct {
	Id    int64
	Name  string
	Phone string
	Etag  uint64
}

type rawRecord struct {
	sqlite_rw.SimpleRow
	*Record
}

func (r *rawRecord) init(bo *Record) *rawRecord {
	r.Record = bo
	return r
}

func (r *rawRecord) Ptrs() []interface{} {
	return []interface{}{&r.Id, &r.Name, &r.Phone}
}

func (r *rawRecord) ValuePtr() interface{} {
	return r.Record
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

func (e *errorRecord) ValuePtr() interface{} {
	return e.Record
}

func (e *errorRecord) Marshall() error {
	return errors.New("An error happened marshalling")
}

func (e *errorRecord) Unmarshall() error {
	return errors.New("An error happened unmarshalling")
}
