package mclitype

import (
	"net/http"
)

type Middleware interface {
	http.Handler
	SetInnerHandler(http.Handler)
}

type CredentialStorer interface {
	GetUser(username string) (Credentialer, error, bool)
	GetUsers(pattern string) (map[string]Credentialer, error)
	SetPassword(username, password string, expired bool) error
	CheckPassword(username, password string) (bool, error)
}

type Credentialer interface {
	SetCredential(username, password string) error
	GetString(field string) (string, error)
}
