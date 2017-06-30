package main

import (
	crand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/pbkdf2"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/russross/meddler"
)

type User struct {
	ID             int64     `meddler:"id,pk"`
	Username       string    `meddler:"username"`
	Admin          bool      `meddler:"admin"`
	Author         bool      `meddler:"author"`
	Password       string    `meddler:"-" json:"password,omitempty"`
	Salt           []byte    `meddler:"salt" json:"-"`
	Scheme         string    `meddler:"scheme" json:"-"`
	PasswordHash   []byte    `meddler:"password_hash" json:"-"`
	LastSignedInAt time.Time `meddler:"last_signed_in_at"`
	CreatedAt      time.Time `meddler:"created_at"`
	ModifiedAt     time.Time `meddler:"modified_at"`
}

func CreateUser(w http.ResponseWriter, r *http.Request, tx *sql.Tx, user User, render render.Render) {
	now := time.Now()
	user.ID = 0

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

	user.Admin = false
	user.Author = false
	user.Scheme = "PBKDF2-HMAC-SHA256:64k"
	user.Salt = make([]byte, 16)
	if _, err := crand.Read(user.Salt); err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "salt error: %v", err)
		return
	}
	start := time.Now()
	user.PasswordHash = pbkdf2.Key(
		[]byte(user.Password),
		user.Salt,
		64*1024,
		32,
		sha256.New)
	log.Printf("hash took %v", time.Since(start))
	user.Password = ""
	user.CreatedAt = now
	user.ModifiedAt = now
	user.LastSignedInAt = now

	// check if the username is in use
	// this is racy, but the unique constraint in the database
	// is the real test. this is just to give a better error message
	var count int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM users WHERE username = ?`, user.Username).Scan(&count); err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "db error: %v", err)
		return
	}
	if count > 0 {
		loggedHTTPErrorf(w, http.StatusBadRequest, "username is already in use")
		return
	}

	if err := meddler.Insert(tx, "users", &user); err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "db error: %v", err)
		return
	}
	render.JSON(http.StatusOK, &user)
}

func GetUsers(w http.ResponseWriter, tx *sql.Tx, render render.Render) {
	users := []*User{}
	if err := meddler.QueryAll(tx, &users, `SELECT * FROM users ORDER BY id`); err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "db error: %v", err)
		return
	}
	render.JSON(http.StatusOK, users)
}

func GetUser(w http.ResponseWriter, tx *sql.Tx, params martini.Params, currentUser *User, render render.Render) {
	userID, err := parseID(w, "user_id", params["user_id"])
	if err != nil {
		return
	}

	user := new(User)
	if err = meddler.Load(tx, "users", user, userID); err != nil {
		loggedHTTPDBNotFoundError(w, err)
		return
	}

	render.JSON(http.StatusOK, user)
}

func GetUserMe(w http.ResponseWriter, currentUser *User, render render.Render) {
	render.JSON(http.StatusOK, currentUser)
}
