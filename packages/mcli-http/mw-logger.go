package mclihttp

import (
	"log"
	"net/http"
)

type logger struct {
	Inner http.Handler
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("start")
	l.Inner.ServeHTTP(w, r)
	log.Println("finish")
}
