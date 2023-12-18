package mcliinterface

import (
	"net/http"

	_ "github.com/Direct-Dev-Ru/go_common_ddru"
)

type ContextKey string

type HandlerFuncsPlugin interface {
	GetHandlerFuncsV2(inputArgs ...interface{}) map[string]http.HandlerFunc
	GetHandlerFuncs()
}
