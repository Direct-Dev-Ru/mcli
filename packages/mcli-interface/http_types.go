package mcliinterface

import "net/http"

type ContextKey string

type HandlerFuncsPlugin interface {
	GetHandlerFuncs() map[string]http.HandlerFunc
}
