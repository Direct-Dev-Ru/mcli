/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>

*/

package main

import (
	"mcli/cmd"
	"os"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	// "github.com/rs/zerolog/log"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	ilogger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	elogger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	elogger.Level(zerolog.ErrorLevel)

	if os.Getenv("DEBUG") == "true" {
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

	iloggers := []zerolog.Logger{ilogger, elogger}
	cmd.Execute(iloggers)
}
