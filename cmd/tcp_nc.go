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
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func nc_1_handle(conn net.Conn) {

	/*
	 * Explicitly calling /bin/sh and using -i for interactive mode
	 * so that we can use it for stdin and stdout.
	 * For Windows use exec.Command("cmd.exe")
	 */
	var cmd *exec.Cmd

	oss := runtime.GOOS
	// fmt.Println(oss)
	switch oss {
	case "windows":
		cmd = exec.Command("cmd.exe")
	// case "darwin":
	// 	cmd = exec.Command("/bin/sh", "-i")
	case "linux":
		cmd = exec.Command("/bin/sh", "-i")
	default:
		cmd = exec.Command("/bin/sh", "-i")
	}

	rp, wp := io.Pipe()
	// Set stdin to our connection
	cmd.Stdin = conn
	cmd.Stdout = wp

	go io.Copy(conn, rp)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	conn.Close()
}

type Flusher struct {
	w *bufio.Writer
}

func NewFlusher(w io.Writer) *Flusher {
	return &Flusher{
		w: bufio.NewWriter(w),
	}
}

func (foo *Flusher) Write(b []byte) (int, error) {
	count, err := foo.w.Write(b)
	if err != nil {
		return -1, err
	}
	if err := foo.w.Flush(); err != nil {
		return -1, err
	}
	return count, err
}

func nc_2_handle(conn net.Conn) {
	defer conn.Close()
	var cmd *exec.Cmd
	oss := runtime.GOOS

	switch oss {
	case "windows":
		cmd = exec.Command("cmd.exe")
	default:
		cmd = exec.Command("/bin/sh", "-i")
	}
	cmd.Stdin = conn

	// cmd.Stderr = os.Stderr
	cmd.Stdout = NewFlusher(conn)

	if err := cmd.Run(); err != nil {
		fmt.Println("while run command error occured :", err.Error())
	} else {

	}

}

func nc_3_handle(conn net.Conn) {

	defer conn.Close()
	reader := bufio.NewReader(conn)
	command, err := reader.ReadString('\n')
	if err != nil {
		Elogger.Error().Msg("unable to read data")
	}
	command = strings.Trim(command, "\r\n")
	command = strings.Trim(command, "\n")
	command = strings.TrimSpace(command)
	if ss := strings.ToLower(command); ss == ":exit" || ss == ":e" || ss == ":quit" || ss == ":q" {
		conn.Write([]byte("nc server stopped"))
		os.Exit(0)
	}
	Ilogger.Trace().Msg(fmt.Sprintf("Read %d bytes: %s", len(command), command))
	Ilogger.Trace().Msg("Process command ...")
	splitted := strings.Split(command, " ")
	executable := splitted[0]
	args := splitted[1:]
	cmd := exec.Command(executable, args...)
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	cmd.Stdout = &outBuf

	if err := cmd.Run(); err != nil {
		errorMessage := "while run command " + command + " error occured :" + err.Error()
		if len(errBuf.Bytes()) > 0 {
			errorMessage += " :: " + errBuf.String()
		}
		Elogger.Error().Msg(errorMessage)
		conn.Write([]byte(errorMessage))

	} else {
		conn.Write(outBuf.Bytes())
	}

}

// ncCmd represents the nc command
var ncCmd = &cobra.Command{
	Use:   "nc",
	Short: "run shell though tcp port",
	Long:  `run shell though tcp port. For example: mcli tcp nc --port 23000`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		host, _ := cmd.Flags().GetString("host")
		Ilogger.Trace().Msg(fmt.Sprintf("Port for nc server is %s", port))

		listener, err := net.Listen("tcp", host+":"+port)
		if err != nil {
			Elogger.Fatal().Msg("unable to bind to port: " + port + " " + err.Error())
		}
		Ilogger.Info().Msg(fmt.Sprintf("Listening on %s:%s", host, port))

		for {
			conn, err := listener.Accept()
			if err != nil {
				Elogger.Fatal().Msg("unable to accept connections: " + port + " " + err.Error())
			}
			go nc_3_handle(conn)
		}
	},
}

func init() {
	tcpCmd.AddCommand(ncCmd)

	// Here you will define your flags and configuration settings.

	ncCmd.Flags().StringP("host", "n", "", "Specify (n)ode")
	ncCmd.Flags().StringP("port", "p", "20001", "Specify (p)ort to listen")
}
