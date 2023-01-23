package mclihttp

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/pkgerrors"
)

type Logger struct {
	infoLog  zerolog.Logger
	errorLog zerolog.Logger
	Inner    http.Handler
}

func NewLogger(inner http.Handler, outInfoLogger zerolog.Logger, outErrorLogger zerolog.Logger) Logger {
	infoLog := outInfoLogger
	errorLog := outErrorLogger

	return Logger{infoLog: infoLog, errorLog: errorLog, Inner: inner}
}

func (l Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.Inner.ServeHTTP(w, r)
	l.infoLog.Trace().Msgf("Request time: %v\n", time.Since(start))
}

func (l Logger) GetHandler(next http.Handler) http.Handler {
	return l.Inner
}
