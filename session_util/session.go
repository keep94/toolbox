// Package session_util provides support for managing web sessions for apps
// where users logs in.
package session_util

import (
  "github.com/gorilla/context"
  "github.com/gorilla/sessions"
  "net/http"
)

// UserIdSession augments a gorilla session by supporting the storing and
// retrieving of the user Id of the logged in user.
type UserIdSession struct {
  S *sessions.Session
}

// UserId returns the userId and true if user Id is stored in this session;
// otherwise it returns 0 and false.
func (s UserIdSession) UserId() (int64, bool) {
  result, ok := s.S.Values[kUserIdKey]
  if !ok {
    return 0, false
  } 
  return result.(int64), true
}

// SetUserId sets the user ID in this session.
func (s UserIdSession) SetUserId(id int64) {
  s.S.Values[kUserIdKey] = id
}

// ClearUserId clears the user ID in this session.
func (s UserIdSession) ClearUserId() {
  delete(s.S.Values, kUserIdKey)
}

// ClearAll clears all data from this session.
func (s UserIdSession) ClearAll() {
  for key := range s.S.Values {
    delete(s.S.Values, key)
  }
}

type UserGetter interface {
  // GetUser retrieves a user from persistent storage given user Id. 
  GetUser(id int64) (userPtr interface{}, err error)
}

// Sessions that store user instances implement this interface.
type UserSession interface {
  // UserId either returns the user id in the session and true or 0 and
  // false if there is no user id in the session.
  UserId() (int64, bool)

  // SetUser sets the user instance in this session.
  SetUser(userPtr interface{})
}

// NewUserSession creates a new UserSession and pairs it with the current
// http request.
// If a user is logged in, the returned UserSession will contain
// that user instance; otherwise returned UserSession will contain
// nil for the user instance. Upon successful completion, caller must call
// context.Clear(r) from github.com/gorilla/context.
// sessionStore is the session store; r is the current http request;
// cookieName is the name of the session cookie;
// factory creates the UserSession given a gorilla session;
// userGetter retrieves user instance from persistent storage given user ID;
// noSuchId is the error that userGetter returns if no such user exist for
// a given ID.
func NewUserSession(
    sessionStore sessions.Store,
    r *http.Request,
    cookieName string,
    factory func(s *sessions.Session) UserSession,
    userGetter UserGetter,
    noSuchId error) (UserSession, error) {
  gs, err := sessionStore.Get(r, cookieName)
  if err != nil {
    return nil, err
  }
  result := factory(gs)
  if userId, ok := result.UserId(); ok {
    userPtr, err := userGetter.GetUser(userId)
    if err == nil {
      result.SetUser(userPtr)
    } else if err != noSuchId {
      return nil, err
    }
  }
  context.Set(r, kSessionContextKey, result)
  return result, nil
}

// GetUserSession returns the UserSession paired with this request. It is
// an error to call GetUserSession on a request without a previously
// successful call to NewUserSession on the same request.
func GetUserSession(r *http.Request) UserSession {
  return context.Get(r, kSessionContextKey).(UserSession)
}

type sessionKeyType int

const (
  kUserIdKey sessionKeyType = iota
)

type contextKeyType int

const (
  kSessionContextKey contextKeyType = iota
)
