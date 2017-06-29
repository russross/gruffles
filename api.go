package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/russross/meddler"
)

func startAPI(db *sql.DB) {
	// set up martini
	r := martini.NewRouter()
	m := martini.New()
	m.Logger(log.New(os.Stderr, "", log.LstdFlags))
	m.User(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	m.Use(render.Renderer(render.Options{IndentJSON: false}))
	m.Use(martini.Static(Config.WWWDir, martini.StaticOptions{SkipLogging: true}))
	store := sessions.NewCookieStore([]byte(Config.SessionSecret))
	store.Options(sessions.Options{
		Path:   "/",
		Secure: true,
		MaxAge: Config.SessionSeconds,
	})
	m.Use(session.Sessions(Config.CookieName, store))

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
	auth := func(w http.ResponseWriter, session sessions.Session) {
		if userID := session.Get("id"); userID == nil {
			loggedHTTPErrorf(w, http.StatusUnauthorized, "authentication: no user ID found in session")
			return
		}
	}

	// martini service: include the current logged-in user (requires withTx and auth)
	withCurrentUser := func(c martini.Context, w http.ResponseWriter, tx *sql.Tx, session sessions.Session) {
		rawID := session.Get("id")
		if rawID == nil {
			loggedHTTPErrorf(w, http.StatusInternalServerError, "cannot find user ID in session")
			return
		}
		userID, ok := rawID.(int64)
		if !ok {
			session.Clear()
			loggedHTTPErrorf(w, http.StatusInternalServerError, "error extracting user ID from session")
			return
		}

		// load the user record
		user := new(User)
		if err := meddler.Load(tx, "users", user, userID); err != nil {
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
			loggedHTTPErrorf(w, http.StatusUnauthorized, "user %d (%s) is not an author", currentUser.ID, currentUser.Name)
			return
		}
	}

	// users
	r.Post("/v1/users", withTx, CreateUser)
	r.Get("/v1/users/me", auth, withTx, withCurrentUser, GetUserMe)
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
