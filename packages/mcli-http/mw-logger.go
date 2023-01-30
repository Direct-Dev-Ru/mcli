package mclihttp

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/pkgerrors"
)

type Logger struct {
	InfoLog  zerolog.Logger
	ErrorLog zerolog.Logger
	ShowUrl  bool
	ShowIp   bool
	Inner    http.Handler
}

type LoggerOpts struct {
	ShowUrl bool
	ShowIp  bool
}

func NewLogger(outILog zerolog.Logger, outErrLog zerolog.Logger, opts LoggerOpts) *Logger {
	return &Logger{InfoLog: outILog, ErrorLog: outErrLog, ShowIp: opts.ShowIp, ShowUrl: opts.ShowUrl}
}

func (l *Logger) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	start := time.Now()
	l.Inner.ServeHTTP(res, req)
	if l.ShowUrl && !l.ShowIp {
		l.InfoLog.Info().Str("URL", req.URL.String()).Msgf("Request time: %v\n", time.Since(start))
	} else if !l.ShowUrl && l.ShowIp {
		l.InfoLog.Info().Str("IP", req.RemoteAddr).Msgf("Request time: %v\n", time.Since(start))
	} else if l.ShowUrl && l.ShowIp {
		l.InfoLog.Info().Str("IP", req.RemoteAddr).Str("URL", req.URL.String()).Msgf("Request time: %v\n", time.Since(start))
	} else {
		l.InfoLog.Info().Msgf("Request time: %v\n", time.Since(start))

	}

	// fmt.Printf("Request time: %v\n", time.Since(start))
}

func (l *Logger) SetInnerHandler(next http.Handler) {
	l.Inner = next
}
