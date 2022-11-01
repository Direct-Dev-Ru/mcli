/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Simple tcp client",
	Long: `Simple tcp client. Sending string message to server. For example:
	supercli tcp client -m "message to server"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		message, _ := cmd.Flags().GetString("message")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")
		timeout, _ := cmd.Flags().GetInt64("timeout")

		Ilogger.Trace().Msg("tcp client called")

		if timeout == 0 {
			timeout = 500
		}
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), time.Duration(timeout*int64(time.Millisecond)))
		if err != nil {
			Elogger.Error().Msg(err.Error())
			return err
		}

		if len(message) > 0 {
			// sending into tcp socket
			writer := bufio.NewWriter(conn)
			if _, err := writer.WriteString(message + "\n"); err != nil {
				Elogger.Fatal().Msg("unable to write data to soket " + host + ":" + port)
			}
			fmt.Println("Message to server: " + message)
			writer.Flush()

			// Listening answer
			messageFromServer, _ := bufio.NewReader(conn).ReadString('\n')
			fmt.Println("Message from server: " + messageFromServer)
			conn.Close()
		} else {
			for {
				// reading from stdin
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Text to send: ")
				s, _ := reader.ReadString('\n')
				// sending into tcp socket
				writer := bufio.NewWriter(conn)
				if _, err := writer.WriteString(s); err != nil {
					Elogger.Fatal().Msg("unable to write data to soket " + host + ":" + port)
				}
				writer.Flush()

				// Listening answer
				messageFromServer, _ := bufio.NewReader(conn).ReadString('\n')
				fmt.Println("Message from server: " + messageFromServer)
				conn.Close()
				fmt.Print("Press c to continue, q to stop session: ")
				cmd, _ := reader.ReadString('\n')
				runeCmd := []rune(cmd)

				if runeCmd[0] != 'q' {

					conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), time.Duration(timeout*int64(time.Millisecond)))
					// Ilogger.Trace().Msg("continue")
					if err != nil {
						Elogger.Error().Msg(err.Error())
						return err
					}
				} else {
					break
				}
			}
		}
		return nil
	},
}

func init() {
	tcpCmd.AddCommand(clientCmd)

	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	clientCmd.Flags().StringP("message", "m", "", "message to send to server")
	clientCmd.Flags().StringP("host", "n", "0.0.0.0", "Specify host/(n)ode for sending messages")
	clientCmd.Flags().StringP("port", "p", "80", "Specify (p)ort for sending messages")
}
