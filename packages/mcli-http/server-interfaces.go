package mclihttp

import (
	"net/http"
)

type Middleware interface {
	http.Handler
	SetInnerHandler(http.Handler)
}
