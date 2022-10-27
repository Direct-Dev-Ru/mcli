/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type scanHostPort struct {
	host   string
	port   int
	result bool
}

type scanParams struct {
	host        string
	netMask     string
	sNetAddress string
	sNetMask    string
	iNetMask    int
	portRange   string
	minPort     int
	maxPort     int
}

func getHostsFromNet(netmask string) (result map[string][]int) {
	result = make(map[string][]int)
	// convert string to IPNet struct "192.168.87.0/27"
	_, ipv4Net, err := net.ParseCIDR(netmask)
	if err != nil {
		return
	}
	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)
	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)
	// loop through addresses as uint32
	for i := start; i < finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		result[ip.String()] = []int{}
		// fmt.Println(ip)
	}
	return
}

func (sp *scanParams) checkScanParameters() error {
	if len(sp.host) == 0 && len(sp.netMask) > 0 {
		rexpAddr, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])/`)
		rexpMask, _ := regexp.Compile(`/[0-9]+$`)
		rexpIpPort := `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])/[0-9]+$`

		matched, err := regexp.MatchString(rexpIpPort, sp.netMask)
		if err != nil {
			return err
		}

		if matched {
			net_addr_slice := rexpAddr.FindAllString(sp.netMask, -1)
			net_mask_slice := rexpMask.FindAllString(sp.netMask, -1)
			error_net_mask := false
			if len(net_addr_slice) > 0 && len(net_mask_slice) > 0 {
				sp.sNetAddress = strings.Trim(net_addr_slice[0], "/")
				sp.sNetMask = strings.Trim(net_mask_slice[0], "/")
				sp.iNetMask, _ = strconv.Atoi(sp.sNetMask)
				error_net_mask = sp.iNetMask > 32

			} else {
				error_net_mask = true
			}
			if error_net_mask {
				return errors.New("netmask have wrong format")
			}
		} else {
			return errors.New("netmask have wrong format")
		}
		rexpPortRange, _ := regexp.Compile(`^[0-9]+-[0-9]+$`)
		if !rexpPortRange.MatchString(sp.portRange) {
			return errors.New("wrong port range specified")
		}
	} else {
		sp.netMask = ""
	}

	ports := strings.Split(sp.portRange, "-")
	sp.minPort, _ = strconv.Atoi(ports[0])
	sp.maxPort, _ = strconv.Atoi(ports[1])

	if sp.minPort > sp.maxPort {
		sp.minPort, _ = strconv.Atoi(ports[1])
		sp.maxPort, _ = strconv.Atoi(ports[0])
	}

	return nil
}

func worker(host string, ports, results chan int) {
	var timeout int64 = 2
	for p := range ports {
		address := fmt.Sprintf("%s:%d", host, p)
		conn, err := net.DialTimeout("tcp", address, time.Duration(timeout*int64(time.Second)))
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}
}

// pscanCmd represents the pscan command
var pscanCmd = &cobra.Command{
	Use:   "pscan -m 192.168.55.1/25 -r 0-1024",
	Short: "Parallel scan port range of given tenwork",
	Long: `Use it to start parallel scan of given network: -m 192.168.55.1/25 -r 0-1024
			For one host set /32 mask.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nmask, _ := cmd.Flags().GetString("netmask")
		host, _ := cmd.Flags().GetString("host")
		prange, _ := cmd.Flags().GetString("portrange")

		scanparams := scanParams{host: host, netMask: nmask,
			portRange: prange}

		// check params for scan
		error := scanparams.checkScanParameters()
		if error != nil {
			return error
		}

		var openports map[string][]int
		// if host specified it is more prioriteted
		if len(scanparams.host) > 0 {
			openports = make(map[string][]int)
			openports[scanparams.host] = []int{}
		} else {
			openports = getHostsFromNet(scanparams.netMask)
			if scanparams.iNetMask == 32 {
				openports[scanparams.sNetAddress] = []int{}
			}
		}
		fmt.Println(scanparams)
		fmt.Println(openports)

		ports := make(chan int, 100)
		results := make(chan int)
		for host, hostPorts := range openports {
			for i := 0; i < cap(ports); i++ {
				go worker(host, ports, results)
			}

			go func() {
				for i := scanparams.minPort; i <= scanparams.maxPort; i++ {
					ports <- i
				}
			}()

			for i := scanparams.minPort; i <= scanparams.maxPort; i++ {
				port := <-results
				if port != 0 {
					openports[host] = append(hostPorts, port)
					fmt.Println(host, port)
				}
			}
		}
		close(ports)
		close(results)
		return nil
	},
}

func init() {
	tcpCmd.AddCommand(pscanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pscanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pscanCmd.Flags().StringP("host", "n", "", "mandatory netmask or host: (-n example.com or -m 192.168.55.0/24)")
	pscanCmd.Flags().StringP("netmask", "m", "", "mandatory netmask or host: (-n example.com or -m 192.168.55.0/24)")
	pscanCmd.Flags().StringP("portrange", "r", "", "mandatory: port range (1:1024)")
}
