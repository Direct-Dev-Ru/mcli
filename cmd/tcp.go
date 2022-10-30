/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// tcpCmd represents the tcp command
var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Quick scan tcp port with parameter",
	Long: `This is a container for a set of subcommands and itself quick scan of port defined by --port parameter
on the host defined by --host parameter. For example:

supercli tcp --host 1.1.1.1 --port 80`,

	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")
		timeout, _ := cmd.Flags().GetInt64("timeout")

		if len(args) > 0 && (len(host) == 0 || len(port) == 0) {
			rexpIp, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):`)
			rexpPort, _ := regexp.Compile(`:[0-9]+$`)
			rexpIpPort := `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):[0-9]+$`

			argIpPort := strings.TrimSpace(args[0])
			matched, err := regexp.MatchString(rexpIpPort, argIpPort)

			if err != nil {
				return err
			}
			if matched {
				host = strings.Trim(rexpIp.FindAllString(argIpPort, -1)[0], ":")
				port = strings.Trim(rexpPort.FindAllString(argIpPort, -1)[0], ":")
			}

		}
		// Scaning host:port
		if host == "" {
			host = "127.0.0.1"
		}
		if port == "" {
			port = "80"
		}

		s := fmt.Sprintf("Now scanning %s:%s ...", host, port)
		fmt.Println(s)
		if timeout == 0 {
			timeout = 500
		}
		Ilogger.Trace().Int64("timeout:", timeout).Send()
		// fmt.Println("timeout", timeout)

		_, errDial := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), time.Duration(timeout*int64(time.Millisecond)))
		if errDial != nil {
			Elogger.Fatal().Msg(fmt.Sprintf("%s:%s error occured: %s", host, port, errDial))
			return nil
		}
		Ilogger.Info().Msg(fmt.Sprintf("%s:%s successfully connected", host, port))
		return nil

	},
}

func init() {
	rootCmd.AddCommand(tcpCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	tcpCmd.PersistentFlags().Int64("timeout", 0, "set timeout for dial in miliseconds")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tcpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// tcpCmd.Flags().String("host", "", "host for probe")
	tcpCmd.Flags().StringP("host", "n", "0.0.0.0", "Specify host/(n)ode for testing connectivity")

	tcpCmd.Flags().StringP("port", "p", "80", "Specify (p)ort for testing connectivity")
	// tcpCmd.Flags().String("port", "", "port for probe")
}
