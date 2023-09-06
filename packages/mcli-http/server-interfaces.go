package mclihttp

import (
	"net/http"
)

type Middleware interface {
	http.Handler
	SetInnerHandler(http.Handler)
}

type CredentialStorer interface {
	GetUser(username string) (*Credential, error, bool)
	GetAllUsers(pattern string) ([]*Credential, error)
	SetPassword(username, password string) error
	CheckPassword(username, password string) (bool, error)
}

// key-value storer interface
