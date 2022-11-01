/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type scanItem struct {
	host   string
	port   int
	result bool
}

func (si scanItem) String() string {
	return fmt.Sprintf("Host %s port %d = %t", si.host, si.port, si.result)
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

func worker(ports, results chan scanItem, timeout int64, id int) {

	for p := range ports {
		resultItem := scanItem{host: p.host, port: p.port, result: false}
		address := fmt.Sprintf("%s:%d", p.host, p.port)
		conn, err := net.DialTimeout("tcp", address, time.Duration(timeout*int64(time.Millisecond)))
		if err != nil {
			// fmt.Println("wId:", id, "fault:", resultItem, err)
			results <- resultItem
			continue
		}
		conn.Close()
		resultItem.result = true
		// fmt.Println("wId:", id, "success:", resultItem)
		results <- resultItem
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

		timeout, _ := cmd.Flags().GetInt64("timeout")
		if timeout == 0 {
			timeout = 100
		}

		scanparams := scanParams{host: host, netMask: nmask, portRange: prange}

		// check params for scan
		error := scanparams.checkScanParameters()
		if error != nil {
			return error
		}

		var openports map[string][]int
		var numProbes int = 0
		// if host is specified then it is more prefered
		if len(scanparams.host) > 0 {
			openports = make(map[string][]int)
			openports[scanparams.host] = []int{}
			numProbes = scanparams.maxPort - scanparams.minPort + 1
		} else {
			openports = getHostsFromNet(scanparams.netMask)
			if scanparams.iNetMask == 32 {
				openports[scanparams.sNetAddress] = []int{}
			}
			numProbes = (scanparams.maxPort - scanparams.minPort + 1) * len(openports)
		}
		fmt.Println(scanparams)

		var workerCount int = 100
		if numProbes < 100 {
			workerCount = numProbes
		}
		ports := make(chan scanItem, workerCount)
		results := make(chan scanItem)

		// opening workers for testing ports
		for i := 0; i < workerCount; i++ {
			go worker(ports, results, timeout, i)
		}

		// iteration through map of hosts
		for host := range openports {

			// Sending to workers in chanel ports
			go func() {
				for i := scanparams.minPort; i <= scanparams.maxPort; i++ {
					sItem := scanItem{host: host, port: i, result: false}
					ports <- sItem
				}
			}()

			for i := scanparams.minPort; i <= scanparams.maxPort; i++ {
				sItem := <-results
				if sItem.result {
					openports[host] = append(openports[host], sItem.port)
					Ilogger.Trace().Msg(sItem.String())
					// fmt.Println(sItem)
				}
			}
		}
		close(ports)
		close(results)
		for _, ports := range openports {
			sort.Ints(ports)
		}
		fmt.Println(openports)
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
