package main

import "time"

type User struct {
	ID             int64     `meddler:"id,pk"`
	Username       string    `meddler:"username"`
	Admin bool `meddler:"admin"`
	Author bool `meddler:"author"`
	Salt           string    `meddler:"salt"`
	Scheme         string    `meddler:"scheme"`
	PasswordHash   string    `meddler:"password_hash"`
	LastSignedInAt time.Time `meddler:"last_signed_in_at"`
	CreatedAt      time.Time `meddler:"created_at"`
	ModifiedAt     time.Time `meddler:"modified_at"`
}
