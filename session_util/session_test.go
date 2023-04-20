package session_util_test

import (
	"errors"
	"fmt"
	"github.com/keep94/context"
	"github.com/keep94/ramstore"
	"github.com/keep94/sessions"
	"github.com/keep94/toolbox/session_util"
	"net/http"
	"strconv"
	"testing"
	"time"
)

const (
	kSessionCookieName = "acookie"
	kSessionId         = "123456"
	kUserId            = 25
)

var (
	kNow = time.Date(2016, 5, 24, 17, 13, 0, 0, time.UTC)
)

var (
	errNoSuchId = errors.New("session_util_test: no such id.")
	errDb       = errors.New("session_util_test: A database error happened.")
)

func TestXsrfToken(t *testing.T) {
	s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
	s.SetUserId(kUserId)
	xsrfToken := s.NewXsrfToken("MyPage", kNow.Add(15*time.Minute))
	if !s.VerifyXsrfToken(xsrfToken, "MyPage", kNow.Add(14*time.Minute)) {
		t.Error("Expected token to verify")
	}
	if s.VerifyXsrfToken(
		xsrfToken, "AnotherPage", kNow.Add(14*time.Minute)) {
		t.Error("Expected token not to verify. Wrong page")
	}
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow.Add(15*time.Minute)) {
		t.Error("Expected token not to verify. Time expired")
	}
}

func TestXsrfTokenUserLogsOut(t *testing.T) {
	s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
	s.SetUserId(kUserId)
	xsrfToken := s.NewXsrfToken("MyPage", kNow.Add(15*time.Minute))
	if !s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token to verify")
	}
	s.ClearUserId()
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token not to verify. User logged out.")
	}
	s.SetUserId(kUserId)
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token not to verify. Secret should have changed.")
	}
}

func TestXsrfTokenClearAll(t *testing.T) {
	s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
	s.SetUserId(kUserId)
	xsrfToken := s.NewXsrfToken("MyPage", kNow.Add(15*time.Minute))
	if !s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token to verify")
	}
	s.ClearAll()
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token not to verify. Session cleared")
	}
	s.SetUserId(kUserId)
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token not to verify. Secret should have changed.")
	}
}

func TestXsrfTokenNewUser(t *testing.T) {
	s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
	s.SetUserId(kUserId)
	xsrfToken := s.NewXsrfToken("MyPage", kNow.Add(15*time.Minute))
	if !s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token to verify")
	}
	s.SetUserId(kUserId + 1)
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token not to verify. Different user.")
	}
	s.SetUserId(kUserId)
	if s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token not to verify. Secret should have changed.")
	}
}

func TestXsrfTokenHack(t *testing.T) {
	s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
	s.SetUserId(kUserId)
	xsrfToken := s.NewXsrfToken("MyPage", kNow.Add(15*time.Minute))
	if !s.VerifyXsrfToken(xsrfToken, "MyPage", kNow) {
		t.Error("Expected token to verify")
	}
	if xsrfToken[10] != ':' {
		t.Error("Expected field dlimiter in xsrf token")
	}
	xsrfExpire := xsrfToken[:10]
	xsrfChecksum := xsrfToken[11:]
	if s.VerifyXsrfToken("", "MyPage", kNow) {
		t.Error("Missing token should not verify.")
	}
	if s.VerifyXsrfToken("garbage", "MyPage", kNow) {
		t.Error("garbage token should not verify.")
	}
	if s.VerifyXsrfToken("garbage:with_field_delimiter", "MyPage", kNow) {
		t.Error("garbage with field delimiter token should not verify.")
	}
	if s.VerifyXsrfToken(
		xsrfExpire+":garbage_checksum", "MyPage", kNow) {
		t.Error("token with garbage checksum should not verify.")
	}
	// Add one to expire in token but leave checksum the same.
	expire, err := strconv.Atoi(xsrfExpire)
	if err != nil {
		t.Errorf("Error happened parsing timestamp %v", err)
	}
	regularToken := fmt.Sprintf("%d:%s", expire, xsrfChecksum)
	hackedToken := fmt.Sprintf("%d:%s", expire+1, xsrfChecksum)
	if !s.VerifyXsrfToken(regularToken, "MyPage", kNow) {
		t.Error("Expected regular token to verify")
	}
	if s.VerifyXsrfToken(hackedToken, "MyPage", kNow) {
		t.Error("Expected hacked token not to verify")
	}
}

func TestSessionUserId(t *testing.T) {
	s := session_util.UserIdSession{&sessions.Session{Values: make(map[interface{}]interface{})}}
	s.SetUserId(kUserId)
	s.SetLastLogin(kNow)
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

	lastLogin, ok := s.LastLogin()
	if !ok {
		t.Error("Expected a last login")
	}
	if lastLogin != kNow {
		t.Errorf("Expected %v, got %v", kNow, lastLogin)
	}

	s.ClearLastLogin()
	_, ok = s.LastLogin()
	if ok {
		t.Error("Did not expect a last login.")
	}

}

func TestSessionClearAll(t *testing.T) {
	m := map[interface{}]interface{}{1: 2, 3: 4}
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
		func(s *sessions.Session) session_util.UserSession {
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
		func(s *sessions.Session) session_util.UserSession {
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
		func(s *sessions.Session) session_util.UserSession {
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
		func(s *sessions.Session) session_util.UserSession {
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
		func(s *sessions.Session) session_util.UserSession {
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
