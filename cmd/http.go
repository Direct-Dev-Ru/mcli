/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mcli_http "mcli/packages/mcli-http"

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
		var port, staticPath, staticPrefix, tlsKey, tlsCert string
		var timeout int64 = 0

		port, _ = cmd.Flags().GetString("port")
		isPortSet := cmd.Flags().Lookup("port").Changed
		staticPath, _ = cmd.Flags().GetString("static-path")
		isStaticPathSet := cmd.Flags().Lookup("static-path").Changed
		staticPrefix, _ = cmd.Flags().GetString("static-prefix")
		isStaticPrefix := cmd.Flags().Lookup("static-prefix").Changed

		tlsKey, _ = cmd.Flags().GetString("tls-key")
		tlsCert, _ = cmd.Flags().GetString("tls-cert")
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

		tmplPath, _ := GetStringParam("tmpl-path", cmd, Config.Http.Server.TmplPath)
		tmplPrefix, _ := GetStringParam("tmpl-prefix", cmd, Config.Http.Server.TmplPrefix)

		// Wait for interrupt signal
		StopHttpChan := make(chan os.Signal, 1)

		// mcli_http.InitMainRoutes(staticPath, staticPrefix)
		r := mcli_http.NewRouter(staticPath, staticPrefix)

		// root route
		rootRoute := mcli_http.NewRoute("/", mcli_http.Equal)
		rootRoute.SetHandler(func(res http.ResponseWriter, req *http.Request) {
			sPath := "./" + staticPath
			if strings.HasPrefix(staticPath, "/") {
				sPath = staticPath
			}
			mainPagePath := ""
			mainPagePathCandidate := sPath + "/index.html"
			if _, err := os.Stat(mainPagePathCandidate); err != nil {
				mainPagePathCandidate = sPath + "/html/index.html"
				if _, err := os.Stat(mainPagePathCandidate); err != nil {
					mainPagePathCandidate = ""
				}
			}
			mainPagePath = mainPagePathCandidate

			if len(mainPagePath) > 0 {
				http.ServeFile(res, req, mainPagePath)
			} else {
				http.Error(res, "404 Not Found Root Index.html", 404)
			}
		})
		r.AddRoute(rootRoute)

		// templates
		tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "/", "/", -1)
		tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "\\", "\\", -1)
		tmplPath = strings.TrimSpace(tmplPath)
		tmplPath = strings.TrimRight(tmplPath, "/")

		if e, _ := IsPathExists(tmplPath); e {
			cache, err := mcli_http.LoadTemplatesCache("http-data/templates")
			if err != nil {
				Elogger.Fatal().Msgf("template caching error: %v", err)
			}
			Ilogger.Trace().Msg(tmplPath + " " + tmplPrefix)
			tmplPrefix = "/" + tmplPrefix + "/"

			tmplRoute := mcli_http.NewRoute(tmplPrefix, mcli_http.Prefix)
			tmplRoute.SetHandler(func(res http.ResponseWriter, req *http.Request) {
				url := req.URL.Path
				tmplName := strings.TrimLeft(url, tmplPrefix)
				tmplKey := tmplPath + "/" + tmplName
				tmpl, ok := cache[tmplKey]
				if !ok {
					http.Error(res, "404 Not Found Template "+tmplKey, 404)
					return
				}
				queryData := req.URL.Query().Get("data")
				err = tmpl.Execute(res, struct {
					Req  *http.Request
					Data interface{}
				}{Req: req, Data: struct{ Dummy string }{"Hello world !!!"}})
				if err != nil {
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
				Ilogger.Trace().Msg(url + " : " + queryData)
			})
			r.AddRoute(tmplRoute)
		}

		r.AddRouteWithHandler("/echo", mcli_http.Prefix, mcli_http.Http_Echo)

		var srv *http.Server

		if len(tlsCert) > 0 && len(tlsKey) > 0 {

			srv = &http.Server{
				Addr:         ":" + port,
				Handler:      r,
				ReadTimeout:  time.Duration(timeout * int64(time.Millisecond)),
				WriteTimeout: time.Duration(timeout * 3 * int64(time.Millisecond)),
				IdleTimeout:  time.Duration(timeout * 4 * int64(time.Millisecond)),
				TLSConfig: &tls.Config{
					MinVersion:               tls.VersionTLS13,
					PreferServerCipherSuites: true,
				},
			}

			go func() {
				if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
					Elogger.Fatal().Msg(err.Error())
				}
			}()
		} else {
			srv = &http.Server{
				Addr:         ":" + port,
				Handler:      r,
				ReadTimeout:  time.Duration(timeout * int64(time.Millisecond)),
				WriteTimeout: time.Duration(timeout * 3 * int64(time.Millisecond)),
				IdleTimeout:  time.Duration(timeout * 4 * int64(time.Millisecond)),
			}

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					Elogger.Fatal().Msg(err.Error())
				}
			}()
		}

		signal.Notify(StopHttpChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		Ilogger.Info().Msg(fmt.Sprintf("http server started on port %s", port))

		<-StopHttpChan
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

	httpCmd.Flags().Int64P("timeout", "t", 5000, "Specify timeout for http server service")

	// setup flags
	var port, staticPath, staticPrefix string = "8080", "http-static", "static"
	var tmplPath, tmplPrefix string = "http-data/templates", "tmpl"

	httpCmd.Flags().StringP("port", "p", port, "Specify port for test http server.")
	httpCmd.Flags().String("static-path", staticPath, "Specify relative path to static folder")
	httpCmd.Flags().String("static-prefix", staticPrefix, "Specify url prefix part to static content")
	httpCmd.Flags().String("tmpl-path", tmplPath, "Specify relative path to template folder")
	httpCmd.Flags().String("tmpl-prefix", tmplPrefix, "Specify url prefix part to handle template content")
	httpCmd.Flags().String("tls-cert", "", "Specify tls-cert file")
	httpCmd.Flags().String("tls-key", "", "Specify tls-key file")
}
