/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

func echo(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	s, err := reader.ReadString('\n')
	if err != nil {
		Elogger.Fatal().Msg("unable to read data")
	}
	Ilogger.Trace().Msg(fmt.Sprintf("Read %d bytes: %s", len(s), s))
	Ilogger.Trace().Msg("Writing data ...")

	writer := bufio.NewWriter(conn)
	if _, err := writer.WriteString(s); err != nil {
		Elogger.Fatal().Msg("unable to write data")
	}
	writer.Flush()

	// or e can use another approach
	// Копируем данные из io.Reader в io.Writer через io.Copy()
	// if _, err := io.Copy(conn, conn); err != nil {
	// 	Elogger.Fatal().Msg("unable to read/write data")
	// }
}

// echoCmd represents the echo command
var echoCmd = &cobra.Command{
	Use:   "echo",
	Short: "Starts simple tcp echo server",
	Long:  `Starts simple tcp echo server. For example: supercli tcp echo -p 33333 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetString("port")
		Ilogger.Trace().Msg(fmt.Sprintf("Port for echo server is %s", port))

		// Binding to tcp port on all interfaces
		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			Elogger.Fatal().Msg("unable to bind to port:" + port)
		}
		Ilogger.Info().Msg(fmt.Sprintf("Listening on 0.0.0.0:%s", port))
		for {
			// Waiting for connection and create net.Conn
			conn, err := listener.Accept()
			Ilogger.Info().Msg(fmt.Sprintf("Received connection on 0.0.0.0:%s", port))
			if err != nil {
				Elogger.Fatal().Msg(fmt.Sprintf("unable to accept connection on 0.0.0.0:%s", port))
			}
			// Process connection using goroutines
			go echo(conn)
		}
		return nil
	},
}

func init() {
	tcpCmd.AddCommand(echoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// echoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	echoCmd.Flags().StringP("port", "p", "20001", "Specify port for echo server")
}
