package mclihttp

import (
	"net/http"
)

type MIddleware interface {
	GetHandler(next http.Handler)
}
