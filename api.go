package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"golang.org/x/crypto/acme/autocert"

	"github.com/go-martini/martini"
	mgzip "github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/russross/meddler"
)

func setupAPI(db *sql.DB) *http.Server {
	// set up martini
	r := martini.NewRouter()
	m := martini.New()
	m.Logger(log.New(os.Stderr, "", log.Lshortfile))
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	m.Use(mgzip.All())
	m.Use(martini.Static(Config.ClientDir, martini.StaticOptions{
		SkipLogging: true,
		Prefix:      "play",
	}))
	m.Use(render.Renderer(render.Options{IndentJSON: true}))

	withTx := func(c martini.Context, w http.ResponseWriter) {
		// start a transaction
		tx, err := db.Begin()
		if err != nil {
			loggedHTTPErrorf(w, http.StatusInternalServerError, "db error starting transaction: %v", err)
			return
		}

		// pass it on to the main handler
		c.Map(tx)
		c.Next()

		// was it a successful result?
		rw := w.(martini.ResponseWriter)
		if rw.Status() < http.StatusBadRequest {
			// commit the transaction
			if err := tx.Commit(); err != nil {
				loggedHTTPErrorf(w, http.StatusInternalServerError, "db error committing transaction: %v", err)
				return
			}
		} else {
			// rollback
			log.Printf("rolling back transaction")
			if err := tx.Rollback(); err != nil {
				loggedHTTPErrorf(w, http.StatusInternalServerError, "db error rolling back transaction: %v", err)
				return
			}
		}
	}

	// martini service: to require an active logged-in session
	auth := func(w http.ResponseWriter, r *http.Request) {
		_, err := GetSession(r)
		if err != nil {
			loggedHTTPErrorf(w, http.StatusUnauthorized, "authentication failed: try logging in again")
			log.Printf("%v", err)
			return
		}
	}

	// martini service: include the current logged-in user (requires withTx and auth)
	withCurrentUser := func(c martini.Context, w http.ResponseWriter, r *http.Request, tx *sql.Tx) {
		session, err := GetSession(r)
		if err != nil {
			loggedHTTPErrorf(w, http.StatusUnauthorized, "authentication failed: try logging in again")
			log.Printf("%v", err)
			return
		}

		// load the user record
		userID := session.UserID
		user := new(User)
		if err := meddler.Load(tx, "users", user, userID); err != nil {
			session.Delete(w)

			if err == sql.ErrNoRows {
				loggedHTTPErrorf(w, http.StatusUnauthorized, "user %d not found", userID)
				return
			}
			loggedHTTPErrorf(w, http.StatusInternalServerError, "db error: %v", err)
			return
		}

		// map the current user to the request context
		c.Map(user)
	}

	administratorOnly := func(w http.ResponseWriter, currentUser *User) {
		if !currentUser.Admin {
			loggedHTTPErrorf(w, http.StatusUnauthorized, "user %d (%s) is not an administrator", currentUser.ID, currentUser.Username)
			return
		}
	}

	// martini service: require logged in user to be an author or administrator (requires withCurrentUser)
	authorOnly := func(w http.ResponseWriter, tx *sql.Tx, currentUser *User) {
		if currentUser.Admin {
			return
		}
		if !currentUser.Author {
			loggedHTTPErrorf(w, http.StatusUnauthorized, "user %d (%s) is not an author", currentUser.ID, currentUser.Username)
			return
		}
	}

	// users
	r.Post("/v1/users", withTx, CreateUser)
	r.Get("/v1/users", auth, withTx, withCurrentUser, administratorOnly, GetUsers)
	r.Get("/v1/users/:user_id", auth, withTx, withCurrentUser, administratorOnly, GetUser)
	r.Get("/v1/users/me", auth, withTx, withCurrentUser, GetUserMe)

	// set up letsencrypt
	lem := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(Config.LetsEncryptCache),
		HostPolicy: autocert.HostWhitelist(Config.Hostname),
		Email:      Config.LetsEncryptEmail,
	}

	// start the https server
	return &http.Server{
		Addr:    ":https",
		Handler: m,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS10,
			GetCertificate:           lem.GetCertificate,
		},
	}
}

func loggedErrorf(f string, params ...interface{}) error {
	log.Print(logPrefix() + fmt.Sprintf(f, params...))
	return fmt.Errorf(f, params...)
}

func loggedHTTPErrorf(w http.ResponseWriter, status int, format string, params ...interface{}) error {
	msg := fmt.Sprintf(format, params...)
	log.Print(logPrefix() + msg)
	http.Error(w, msg, status)
	return fmt.Errorf("%s", msg)
}

func loggedHTTPDBNotFoundError(w http.ResponseWriter, err error) {
	msg := "not found"
	status := http.StatusNotFound
	if err != sql.ErrNoRows {
		msg = fmt.Sprintf("db error: %v", err)
		status = http.StatusInternalServerError
	}
	log.Print(logPrefix(), msg)
	http.Error(w, msg, status)
}

func logPrefix() string {
	prefix := ""
	if _, file, line, ok := runtime.Caller(2); ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		prefix = fmt.Sprintf("%s:%d: ", file, line)
	}
	return prefix
}

func unBase64(s string) string {
	if raw, err := base64.StdEncoding.DecodeString(s); err == nil {
		return string(raw)
	}
	return s
}
