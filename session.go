package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/pbkdf2"

	"github.com/gorilla/securecookie"
	"github.com/martini-contrib/render"
	"github.com/russross/meddler"
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

func CreateSession(w http.ResponseWriter, r *http.Request, tx *sql.Tx, user User, render render.Render) {
	now := time.Now()

	// username: letters, digits, underscores, hyphens, max 32 characters
	if !utf8.ValidString(user.Username) {
		loggedHTTPErrorf(w, http.StatusBadRequest, "username must be valid utf-8")
		return
	}
	user.Username = strings.TrimSpace(user.Username)
	user.Username = strings.ToLower(user.Username)
	if len(user.Username) < 1 || len(user.Username) > 32 {
		loggedHTTPErrorf(w, http.StatusBadRequest, "username must be between 1 and 32 characters")
		return
	}
	for _, ch := range user.Username {
		if ch <= ' ' || ch > '~' {
			loggedHTTPErrorf(w, http.StatusBadRequest, "username can contain only printable ASCII characters")
			return
		}
	}

	// password must be between 12 and 256 characters
	if len(user.Password) < 12 || len(user.Password) > 256 {
		loggedHTTPErrorf(w, http.StatusBadRequest, "password must be between 12 and 256 characters")
		return
	}

	realUser := new(User)
	if err := meddler.QueryRow(tx, realUser, `SELECT * FROM users WHERE username = ?`, user.Username); err != nil {
		if err == sql.ErrNoRows {
			time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
			loggedHTTPErrorf(w, http.StatusUnauthorized, "no such user")
			return
		}
		loggedHTTPErrorf(w, http.StatusInternalServerError, "db error: %v", err)
		return
	}
	if realUser.Scheme != "PBKDF2-HMAC-SHA256:64k" {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "unknown hash scheme")
		return
	}
	start := time.Now()
	hash := pbkdf2.Key(
		[]byte(user.Password),
		realUser.Salt,
		64*1024,
		32,
		sha256.New)
	log.Printf("hash took %v", time.Since(start))
	user.Password = ""
	if subtle.ConstantTimeCompare(hash, realUser.PasswordHash) != 1 {
		time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
		loggedHTTPErrorf(w, http.StatusUnauthorized, "wrong password")
		return
	}
	realUser.LastSignedInAt = now

	if err := meddler.Update(tx, "users", realUser); err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "db error: %v", err)
		return
	}

	// form a session
	session, err := NewSession(r, realUser.ID)
	if err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "session error: %v", err)
		return
	}
	session.Save(w)
	render.JSON(http.StatusOK, session)
}
