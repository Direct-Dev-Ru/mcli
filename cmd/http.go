/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"fmt"
	"net/http"
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
		// process configuration or setup defaults
		if len(Config.Http.Server.Port) > 0 {
			port = Config.Http.Server.Port
		} else {
			port, _ = cmd.Flags().GetString("port")
		}

		if len(Config.Http.Server.StaticPath) > 0 {
			staticPath = Config.Http.Server.StaticPath
		} else {
			staticPath, _ = cmd.Flags().GetString("static-path")
		}

		if len(Config.Http.Server.StaticPrefix) > 0 {
			staticPrefix = Config.Http.Server.StaticPrefix
		} else {
			staticPrefix, _ = cmd.Flags().GetString("static-prefix")
		}

		Ilogger.Trace().Msg(fmt.Sprintf("Port for http server is %s", port))
		go func() {
			time.Sleep(time.Second * 2)
			Ilogger.Info().Msg(fmt.Sprintf("http server started on port %s", port))
		}()

		mclihttp.InitMainRoutes(staticPath, staticPrefix)
		err := http.ListenAndServe(":"+port, nil)

		if err != nil {
			Elogger.Fatal().Msg("unable to bind to port: " + port + " " + err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	// httpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// setup flags
	var port, staticPath, staticPrefix string = "8080", "http-static", "static"

	httpCmd.Flags().StringP("port", "p", port, "Specify port for test http server.")
	httpCmd.Flags().String("static-path", staticPath, "Specify relative part of path to static folder")
	httpCmd.Flags().String("static-prefix", staticPrefix, "Specify url prefix part to static folder")
}
