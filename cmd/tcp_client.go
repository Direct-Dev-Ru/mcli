/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Simple tcp client",
	Long: `Simple tcp client. Sending string message to server. For example:
	mcli tcp client -m "message to server"
`,
	Run: func(cmd *cobra.Command, args []string) {
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
			Elogger.Fatal().Msg(err.Error())
		}
		defer conn.Close()

		if len(message) > 0 {
			// sending into tcp socket
			writer := bufio.NewWriter(conn)
			if _, err := writer.WriteString(message + "\n"); err != nil {
				Elogger.Fatal().Msg("unable to write data to socket " + host + ":" + port)
			}
			fmt.Println("Message to server: ")
			fmt.Println(message)
			writer.Flush()

			// Listening answer

			// mes, err := ioutil.ReadAll(conn)
			var outBuf bytes.Buffer
			// mes := make([]byte, 4096, 4096)
			_, err := io.Copy(&outBuf, conn)
			// _, err := conn.Read(outBuf.Bytes())
			if err != nil && err != io.EOF {
				Elogger.Fatal().Msg(err.Error())
			}
			// outServer := string(mes)
			outServer := outBuf.String()
			outServer = strings.Trim(outServer, "\x00")
			outServer = strings.Trim(outServer, "\r\n")
			outServer = strings.Trim(outServer, "\n")
			fmt.Println("Message from server: ")
			fmt.Println(outServer)

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
					Elogger.Fatal().Msg("unable to write data to socket " + host + ":" + port)
				}
				writer.Flush()

				// Listening answer
				// mes := make([]byte, 4096, 4096)
				// _, err := conn.Read(mes)
				var outBuf bytes.Buffer
				// mes := make([]byte, 4096, 4096)
				_, err := io.Copy(&outBuf, conn)

				if err != nil && err != io.EOF {
					Elogger.Fatal().Msg(err.Error())
				}
				// outServer := string(mes)
				outServer := outBuf.String()
				outServer = strings.Trim(outServer, "\x00")
				outServer = strings.Trim(outServer, "\r\n")
				outServer = strings.Trim(outServer, "\n")
				fmt.Println("Message from server: ")
				fmt.Println(outServer)
				conn.Close()
				if outServer == "nc server stopped" {
					os.Exit(0)
				}
				// asking user to quit or continue
				fmt.Print("Press c to continue, q to stop session: ")
				cmd, _ := reader.ReadString('\n')
				runeCmd := []rune(cmd)

				if runeCmd[0] != 'q' {
					conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), time.Duration(timeout*int64(time.Millisecond)))
					// Ilogger.Trace().Msg("continue")
					if err != nil {
						Elogger.Fatal().Msg(err.Error())
					}
				} else {
					break
				}
			}
		}
	},
}

func init() {
	tcpCmd.AddCommand(clientCmd)

	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	clientCmd.Flags().StringP("message", "m", "", "message to send to server")
	clientCmd.Flags().StringP("host", "n", "0.0.0.0", "Specify host/(n)ode for sending messages")
	clientCmd.Flags().StringP("port", "p", "80", "Specify (p)ort for sending messages")
}
