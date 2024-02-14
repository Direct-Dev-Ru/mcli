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
	// "github.com/rs/zerolog/log"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	iLogger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	eLogger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	eLogger.Level(zerolog.ErrorLevel)

	ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
	
	if ENV_DEBUG == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		buildInfo, _ := debug.ReadBuildInfo()

		iLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			Level(zerolog.TraceLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Str("go_version", buildInfo.GoVersion).
			Logger()

		eLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			Level(zerolog.WarnLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Str("go_version", buildInfo.GoVersion).
			Logger()
	}

	// ilogger.Info().Msg("Hello from Zerolog global ilogger")

	// Elogger.Error().
	// 	Stack().
	// 	Err(errors.New("file open failed!")).
	// 	Msg("something happened!")

	iloggers := []zerolog.Logger{iLogger, eLogger}
	cmd.Execute(iloggers)
}
