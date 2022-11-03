/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"net"
	"strconv"

	"github.com/spf13/cobra"
)

func proxy_handle(srcConn net.Conn, dst string) {
	// ❶
	dstConn, err := net.Dial("tcp", dst)
	if err != nil {
		Elogger.Fatal().Msg("unable to connect to host " + dst + " " + err.Error())
	}
	defer dstConn.Close()

	//❷ we exucute code in goroutine to prevent blocking io.Copy
	go func() {
		// Copying from src to dst ❸

		if w, err := io.Copy(dstConn, srcConn); err != nil {
			Elogger.Fatal().Msg("from src to dest copying: " + err.Error())
		} else {
			Ilogger.Trace().Msg("Written from src to dest : " + strconv.Itoa(int(w)))
		}

	}()
	//Copying from dst to src ❹
	if w, err := io.Copy(srcConn, dstConn); err != nil {
		Elogger.Fatal().Msg("from dest to src copying: " + err.Error())
	} else {
		Ilogger.Trace().Msg("Written from dest to src : " + strconv.Itoa(int(w)))
	}
}

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Simple tcp proxy command",
	Long: ` Simple tcp proxy to rewrite all from one host:port to another. For example: mcli tcp proxy -d 127.0.0.1:8080 -s 127.0.0.1:80

`,
	Run: func(cmd *cobra.Command, args []string) {
		dst, _ := cmd.Flags().GetString("dst")
		src, _ := cmd.Flags().GetString("src")
		// timeout, _ := cmd.Flags().GetInt64("timeout")

		// listening for src host:port
		listener, err := net.Listen("tcp", src)
		if err != nil {
			Elogger.Fatal().Msg("unable to bind to port: " + src + " " + err.Error())
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				Elogger.Fatal().Msg("unable to accept connections: " + src + " " + err.Error())
			}
			go proxy_handle(conn, dst)
		}

	},
}

func init() {
	tcpCmd.AddCommand(proxyCmd)

	// Here you will define your flags and configuration settings.

	// proxyCmd.PersistentFlags().String("foo", "", "A help for foo")

	proxyCmd.Flags().StringP("dst", "d", "127.0.0.1:8080", "Specify (d)estination host:port for proxying")

	proxyCmd.Flags().StringP("src", "s", ":80", "Specify (s)ource host:port for proxying")

}
