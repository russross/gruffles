package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type Session struct {
	UserID       int64
	SignedInFrom string
	SignedInAt   time.Time
	ExpiresAt    time.Time
	path         string
}

func NewSession(r *http.Request, userID int64) (*Session, error) {
	now := time.Now()
	client := r.Header.Get("X-Forwarded-For")
	if client == "" {
		client = r.Header.Get("X-Real-IP")
	}
	if client == "" {
		client = r.RemoteAddr
		host, _, err := net.SplitHostPort(client)
		if err != nil {
			return nil, fmt.Errorf("finding client address: %v", err)
		}
		client = host
	}
	expires := now.Add(time.Duration(Config.SessionSeconds) * time.Second)
	return &Session{
		UserID:       userID,
		SignedInFrom: client,
		SignedInAt:   now,
		ExpiresAt:    now.Add(time.Duration(Config.SessionSeconds) * time.Second),
		path:         "/v1/",
	}, nil
}

func GetSession(r *http.Request) (*Session, error) {
	now := time.Now()

	cookie, err := r.Cookie(Config.CookieName)
	if err != nil {
		return nil, fmt.Errorf("unable to read session cookie")
	}

	// decode and verify signature
	session := new(Session)
	secure := securecookie.New([]byte(Config.SessionSecret), nil)
	secure.MaxAge(0)
	if err = secure.Decode(Config.CookieName, cookie.Value, session); err != nil {
		return nil, fmt.Errorf("unable to decode session cookie")
	}

	// check expiration
	if session.ExpiresAt.Before(now) {
		return nil, fmt.Errorf("session is expired; must log in again to continue")
	}

	// sanity check
	if session.UserID < 1 {
		return nil, fmt.Errorf("session does not contain a legal user ID field")
	}

	return session, nil
}

func (session *Session) Save(w http.ResponseWriter) {
	// encode and sign
	secure := securecookie.New([]byte(Config.SessionSecret), nil)
	secure.MaxAge(0)
	encoded, err := secure.Encode(Config.CookieName, session)
	if err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "creating session: %v", err)
		return
	}

	cookie := &http.Cookie{
		Name:    Config.CookieName,
		Value:   encoded,
		Path:    session.path,
		Expires: session.ExpiresAt,
		MaxAge:  int(time.Until(session.ExpiresAt).Seconds()),
		Secure:  true,
	}
	http.SetCookie(w, cookie)
}

func (session *Session) Delete(w http.ResponseWriter) {
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	cookie := &http.Cookie{
		Name:    Config.CookieName,
		Value:   "deleted",
		Path:    session.path,
		Expires: epoch,
		MaxAge:  -1,
		Secure:  true,
	}
	http.SetCookie(w, cookie)
}
