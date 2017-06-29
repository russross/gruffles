package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/pbkdf2"

	"github.com/martini-contrib/render"
	"github.com/russross/meddler"
)

type User struct {
	ID             int64     `meddler:"id,pk"`
	Username       string    `meddler:"username"`
	Admin          bool      `meddler:"admin"`
	Author         bool      `meddler:"author"`
	Password       *string   `meddler:"-" json:"password,omitempty"`
	Salt           string    `meddler:"salt" json:"-"`
	Scheme         string    `meddler:"scheme" json:"-"`
	PasswordHash   string    `meddler:"password_hash" json:"-"`
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
	if len(user.Username) == 0 || len(user.Username) > 32 {
		loggedHTTPErrorf(w, http.StatusBadRequest, "username must be between 1 and 32 characters")
		return
	}
	for _, ch := range user.Username {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') && ch != '_' {
			loggedHTTPErrorf(w, http.StatusBadRequest, "username can contain only letters, digits, and underscores")
			return
		}
	}

	// password must be between 12 and 256 characters
	if user.Password == nil || len(*user.Password) < 12 || len(*user.Password) > 256 {
		loggedHTTPErrorf(w, http.StatusBadRequest, "password must be between 12 and 256 characters")
		return
	}

	user.Admin = false
	user.Author = false
	user.Scheme = "PBKDF2-HMAC-SHA256:64k"
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		loggedHTTPErrorf(w, http.StatusInternalServerError, "salt error: %v", err)
		return
	}
	user.Salt = base64.StdEncoding.EncodeToString(salt)
	hash := pbkdf2.Key([]byte(*user.Password), salt, 65536, 32, sha256.New)
	user.PasswordHash = base64.StdEncoding.EncodeToString(hash)
	user.Password = nil
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
