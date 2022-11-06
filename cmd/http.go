/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"fmt"
	"net/http"
	"time"

	_ "mcli/packages/mcli-http"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts simple http server for testing purposes and it is top level command for other http servers and services",
	Long: `Starts simple http server for testing purposes and it is top level command for other http servers and services
	For example: mcli http
`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		Ilogger.Trace().Msg(fmt.Sprintf("Port for http server is %s", port))
		go func() {
			time.Sleep(time.Second * 2)
			Ilogger.Trace().Msg(fmt.Sprintf("Test Http server started on port %s", port))
		}()
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			Elogger.Fatal().Msg("unable to bind to port: " + port + " " + err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	// httpCmd.PersistentFlags().String("foo", "", "A help for foo")

	httpCmd.Flags().StringP("port", "p", "8080", "Specify port for test http server.")
}
