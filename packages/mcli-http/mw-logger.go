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

func NewLogger(outInfoLogger zerolog.Logger, outErrorLogger zerolog.Logger) *Logger {
	infoLog := outInfoLogger
	errorLog := outErrorLogger

	return &Logger{infoLog: infoLog, errorLog: errorLog}
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.Inner.ServeHTTP(w, r)
	l.infoLog.Trace().Msgf("Request time: %v\n", time.Since(start))
	// fmt.Printf("Request time: %v\n", time.Since(start))
}

func (l *Logger) SetInnerHandler(next http.Handler) {
	l.Inner = next
}
