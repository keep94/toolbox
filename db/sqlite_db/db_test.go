package sqlite_db

import (
  "errors"
  "fmt"
  "github.com/keep94/gosqlite/sqlite"
  "sync"
  "testing"
)

var (
  beginFailed = errors.New("sqlite_db: begin failed.")
  commitFailed = errors.New("sqlite_db: commit failed.")
)

func TestTheDb(t *testing.T) {
  var wg1, wg2 sync.WaitGroup
  conn, _ := sqlite.Open(":memory:")
  db := New(conn)
  wg1.Add(2)
  wg2.Add(1)
  go func() {
    runAndStop(t, db, 0, 100)
    wg1.Done()
  }()
  go func() {
    runAndStop(t, db, 100, 200)
    wg1.Done()
  }()
  go func() {
    runForever(t, db)
    wg2.Done()
  }()
  wg1.Wait()
  if output := db.Close(); output != nil {
    t.Errorf("Expected nil got %v", output)
  }
  if output := conn.Close(); output == nil {
    t.Error("Expected connection to be closed.")
  }
  wg2.Wait()
}

func TestCommit(t *testing.T) {
  wrapper := &fakeConnWrapper{}
  db := new(wrapper)
  if output := db.Do(testActionSucceed); output != nil {
    t.Errorf("Expected nil, got %v", output)
  }
  if !wrapper.verifyCommitted() {
    t.Error("Connection not committed.")
  }
  db.Close()
}

func TestRollback(t *testing.T) {
  wrapper := &fakeConnWrapper{}
  db := new(wrapper)
  if db.Do(testAction(0)) == nil {
    t.Error("Expected non-nil result.")
  }
  if !wrapper.verifyRolledBack() {
    t.Error("Connection not rolled back.")
  }
  db.Close()
}

func TestCommitFailed(t *testing.T) {
  wrapper := &fakeConnWrapper{commitFailure: true}
  db := new(wrapper)
  if output := db.Do(testActionSucceed); output != commitFailed {
    t.Errorf("Expected commitFailed, got %v", output)
  }
  if !wrapper.verifyRolledBack() {
    t.Error("Connection not rolled back.")
  }
  db.Close()
}

func TestBeginFailed(t *testing.T) {
  wrapper := &fakeConnWrapper{beginFailure: true}
  db := new(wrapper)
  if output := db.Do(testActionSucceed); output != beginFailed {
    t.Errorf("Expected beginFailed, got %v", output)
  }
  db.Close()
}

type testError int

func (te testError) Error() string {
  return fmt.Sprintf("%d", te)
}

type fakeConnWrapper struct {
  beginFailure bool
  commitFailure bool
  idx int
  beginCalled int
  commitCalled int
  rollbackCalled int
  delegateCalled int
}

func (f *fakeConnWrapper) begin() error {
  f.idx++
  f.beginCalled = f.idx
  if f.beginFailure {
    return beginFailed
  }
  return nil
}

func (f *fakeConnWrapper) commit() error {
  f.idx++
  f.commitCalled = f.idx
  if f.commitFailure {
    return commitFailed
  }
  return nil
}

func (f *fakeConnWrapper) rollback() error {
  f.idx++
  f.rollbackCalled = f.idx
  return nil
}

func (f *fakeConnWrapper) delegate() *sqlite.Conn {
  f.idx++
  f.delegateCalled = f.idx
  return nil
}

func (f *fakeConnWrapper) verifyCommitted() bool {
  return f.beginCalled > 0 && f.delegateCalled > f.beginCalled && f.commitCalled > f.delegateCalled && f.rollbackCalled == 0
}

func (f *fakeConnWrapper) verifyRolledBack() bool {
  return f.beginCalled > 0 && f.delegateCalled > f.beginCalled && f.rollbackCalled > f.delegateCalled && f.rollbackCalled > f.commitCalled
}

func (f *fakeConnWrapper) Close() error {
  return nil
}
  
func runAndStop(t *testing.T, db *Db, start int, end int) {
  for i := start; i < end; i++ {
    if output := int(db.Do(testAction(i)).(testError)); output != i {
      t.Errorf("Expected %v, got %v", i, output)
      return
    }
  }
}

func runForever(t *testing.T, db *Db) {
  output := db.Do(testActionSucceed)
  for ; output != AlreadyClosed; output = db.Do(testActionSucceed) {
    if output != nil {
      t.Errorf("Expected nil, got %v", output)
      return
    }
  }
}

func testAction(i int) Action {
  return func(conn *sqlite.Conn) error {
    return testError(i)
  }
}

func testActionSucceed(conn *sqlite.Conn) error {
  return nil
}
