package session_util_test

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/session_util"
  "github.com/gorilla/context"
  "github.com/gorilla/sessions"
  "github.com/keep94/ramstore"
  "net/http"
  "testing"
)

const (
  kSessionCookieName = "acookie"
  kSessionId = "123456"
  kUserId = 25
)

var (
  errNoSuchId = errors.New("session_util_test: no such id.")
  errDb = errors.New("session_util_test: A database error happened.")
)

func TestSessionUserId(t *testing.T) {
  s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
  s.SetUserId(kUserId)
  id, ok := s.UserId()
  if !ok {
    t.Error("Expected a UserId")
  }
  if id != kUserId {
    t.Errorf("Expected %d, got %d", kUserId, id)
  }
  s.ClearUserId()
  id, ok = s.UserId()
  if ok {
    t.Error("Did not expect a user Id.")
  }
}

func TestSessionClearAll(t *testing.T) {
  m := map[interface{}]interface{} {1:2, 3:4}
  s := session_util.UserIdSession{&sessions.Session{Values: m}}
  if len(m) != 2 {
    t.Fatal("Expected 2 things in map")
  }
  s.ClearAll()
  if len(m) != 0 {
    t.Error("Expected map to be empty")
  }
}

func TestUserSession(t *testing.T) {
  userStore := store{kUserId}
  sessionStore := newSessionStoreWithUserId(kSessionId, kUserId)
  r := requestWithCookie(kSessionCookieName, kSessionId)
  us, err := session_util.NewUserSession(
      sessionStore,
      r,
      kSessionCookieName,
      func (s *sessions.Session) session_util.UserSession {
        return newUserSession(s)
      },
      userStore,
      errNoSuchId)
  if err != nil {
    t.Fatalf("An error happened getting userSession: %v", err)
  }
  defer context.Clear(r)
  myUserSession := us.(*userSession)
  if output := myUserSession.User; *output != kUserId {
    t.Errorf("Expected %v, got %v", kUserId, *output)
  }
  if myUserSession != session_util.GetUserSession(r) {
    t.Error("User session not stored with request.")
  }
}

func TestUserSessionNoSuchId(t *testing.T) {
  userStore := store{kUserId + 1}
  sessionStore := newSessionStoreWithUserId(kSessionId, kUserId)
  r := requestWithCookie(kSessionCookieName, kSessionId)
  us, err := session_util.NewUserSession(
      sessionStore,
      r,
      kSessionCookieName,
      func (s *sessions.Session) session_util.UserSession {
        return newUserSession(s)
      },
      userStore,
      errNoSuchId)
  if err != nil {
    t.Fatalf("An error happened getting userSession: %v", err)
  }
  defer context.Clear(r)
  myUserSession := us.(*userSession)
  if output := myUserSession.User; output != nil {
    t.Error("Should not have user instance in user session.")
  }
}

func TestUserSessionExpired(t *testing.T) {
  userStore := errorStore{}
  sessionStore := ramstore.NewRAMStore(900)
  r := requestWithCookie(kSessionCookieName, kSessionId)
  us, err := session_util.NewUserSession(
      sessionStore,
      r,
      kSessionCookieName,
      func (s *sessions.Session) session_util.UserSession {
        return newUserSession(s)
      },
      userStore,
      errNoSuchId)
  if err != nil {
    t.Fatalf("An error happened getting userSession: %v", err)
  }
  defer context.Clear(r)
  myUserSession := us.(*userSession)
  if output := myUserSession.User; output != nil {
    t.Error("Should not have user instance in user session.")
  }
}

func TestUserSessionNewSession(t *testing.T) {
  userStore := errorStore{}
  sessionStore := ramstore.NewRAMStore(900)
  r := &http.Request{}
  us, err := session_util.NewUserSession(
      sessionStore,
      r,
      kSessionCookieName,
      func (s *sessions.Session) session_util.UserSession {
        return newUserSession(s)
      },
      userStore,
      errNoSuchId)
  if err != nil {
    t.Fatalf("An error happened getting userSession: %v", err)
  }
  defer context.Clear(r)
  myUserSession := us.(*userSession)
  if output := myUserSession.User; output != nil {
    t.Error("Should not have user instance in user session.")
  }
}

func TestUserSessionError(t *testing.T) {
  userStore := errorStore{}
  sessionStore := newSessionStoreWithUserId(kSessionId, kUserId)
  r := requestWithCookie(kSessionCookieName, kSessionId)
  _, err := session_util.NewUserSession(
      sessionStore,
      r,
      kSessionCookieName,
      func (s *sessions.Session) session_util.UserSession {
        return newUserSession(s)
      },
      userStore,
      errNoSuchId)
  if err != errDb {
    t.Error("Expected to get an error getting user session.")
  }
}

type userSession struct {
  session_util.UserIdSession
  User *int64
}

func newUserSession(s *sessions.Session) *userSession {
  return &userSession{UserIdSession: session_util.UserIdSession{s}}
}

func (u *userSession) SetUser(userPtr interface{}) {
  u.User = userPtr.(*int64)
}

type store struct {
  user int64
}

func (s store) GetUser(id int64) (interface{}, error) {
  if id == s.user {
    return &s.user, nil
  }
  return nil, errNoSuchId
}

type errorStore struct {
}

func (s errorStore) GetUser(id int64) (interface{}, error) {
  return nil, errDb
}

func requestWithCookie(cookieName, cookieValue string) *http.Request {
  cookieHeader := fmt.Sprintf("%s=%s", cookieName, cookieValue)
  return &http.Request{Header: http.Header{"Cookie": {cookieHeader}}}
}

func newSessionStoreWithUserId(sessionId string, userId int64) sessions.Store {
  result := ramstore.NewRAMStore(900)
  sessionData := make(map[interface{}]interface{})
  s := session_util.UserIdSession{&sessions.Session{Values: sessionData}}
  s.SetUserId(userId)
  result.Data.Save(sessionId, sessionData)
  return result
}



