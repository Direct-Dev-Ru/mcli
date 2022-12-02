/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>

*/

package main

import (
	"mcli/cmd"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/term"
	// "github.com/rs/zerolog/log"
)

var TermWidth, TermHeight int = 0, 0
var IsTerminal bool = false

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	ilogger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	elogger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	elogger.Level(zerolog.ErrorLevel)
	ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
	if ENV_DEBUG == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		buildInfo, _ := debug.ReadBuildInfo()

		ilogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			Level(zerolog.TraceLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Str("go_version", buildInfo.GoVersion).
			Logger()

		elogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			Level(zerolog.WarnLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Str("go_version", buildInfo.GoVersion).
			Logger()
	}

	// ilogger.Info().Msg("Hello from Zerolog global ilogger")

	// elogger.Error().
	// 	Stack().
	// 	Err(errors.New("file open failed!")).
	// 	Msg("something happened!")

	if term.IsTerminal(0) {
		// println("in a term")
		IsTerminal = true
	}
	var err error
	TermWidth, TermHeight, err = term.GetSize(0)
	if err != nil {
		IsTerminal = false
	}
	// println("width:", TermWidth, "height:", TermHeight)

	iloggers := []zerolog.Logger{ilogger, elogger}
	cmd.Execute(iloggers)
}
