/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mclihttp "mcli/packages/mcli-http"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts simple http server for testing purposes and it is top level command for other http servers and services",
	Long: `Starts simple http server for testing purposes and it is top level command for other http servers and services
	For example: mcli http --static-path ./http-static --static-prefix static
	it lookup index.html in ./http-static/html/index.html, other static assets in ./http-static/...
	and we can refer to it in url by /static/... prefix
`,
	Run: func(cmd *cobra.Command, args []string) {
		var port, staticPath, staticPrefix string
		var timeout int64 = 0

		port, _ = cmd.Flags().GetString("port")
		isPortSet := cmd.Flags().Lookup("port").Changed
		staticPath, _ = cmd.Flags().GetString("static-path")
		isStaticPathSet := cmd.Flags().Lookup("static-path").Changed
		staticPrefix, _ = cmd.Flags().GetString("static-prefix")
		isStaticPrefix := cmd.Flags().Lookup("static-prefix").Changed
		timeout, _ = cmd.Flags().GetInt64("timeout")
		isTimeoutSet := cmd.Flags().Lookup("timeout").Changed

		// process configuration or setup defaults
		if !isPortSet && len(Config.Http.Server.Port) > 0 {
			port = Config.Http.Server.Port
		}
		if !isTimeoutSet && Config.Http.Server.Timeout > 0 {
			timeout = Config.Http.Server.Timeout
		}

		if !isStaticPathSet && len(Config.Http.Server.StaticPath) > 0 {
			staticPath = Config.Http.Server.StaticPath
		}

		if !isStaticPrefix && len(Config.Http.Server.StaticPrefix) > 0 {
			staticPrefix = Config.Http.Server.StaticPrefix
		}

		mclihttp.InitMainRoutes(staticPath, staticPrefix)
		srv := &http.Server{
			Addr:         ":" + port,
			Handler:      http.DefaultServeMux,
			ReadTimeout:  time.Duration(timeout * int64(time.Millisecond)),
			WriteTimeout: time.Duration(timeout * int64(time.Millisecond)),
		}

		// err := http.ListenAndServe(":"+port, nil)
		// var srvErr error
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				Elogger.Fatal().Msg(err.Error())
			}
		}()

		// Wait for interrupt signal
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		Ilogger.Info().Msg(fmt.Sprintf("http server started on port %s", port))

		<-done
		Ilogger.Info().Msg(fmt.Sprintf("http server stopeed on port %s", port))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			// extra handling here
			cancel()
		}()

		if err := srv.Shutdown(ctx); err != nil {
			Elogger.Fatal().Msg(fmt.Sprintf("server shutdown failed:%+v", err))
		}
		Ilogger.Info().Msg("server exited properly")
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().Int64P("timeout", "t", 5000, "Specify timeout for http services (server and request)")

	// setup flags
	var port, staticPath, staticPrefix string = "8080", "http-static", "static"

	httpCmd.Flags().StringP("port", "p", port, "Specify port for test http server.")
	httpCmd.Flags().String("static-path", staticPath, "Specify relative part of path to static folder")
	httpCmd.Flags().String("static-prefix", staticPrefix, "Specify url prefix part to static folder")
}
