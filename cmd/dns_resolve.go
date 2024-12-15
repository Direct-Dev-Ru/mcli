/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	mcli_utils "mcli/packages/mcli-utils"

	dns "github.com/miekg/dns"
	"github.com/spf13/cobra"
)

type DnsResolveOutData struct {
	HostName    string   `json:"dns-name"`
	IpAdresses  []net.IP `json:"ip-adresses"`
	DnsResolver string   `json:"dns-resolver"`
}

type result struct {
	IPAddress string
	Hostname  string
}

type empty struct{}

func getDnsServerWithPort(serverAddr string) string {
	return fmt.Sprintf("%s:53", serverAddr)
}

func DnsWorker(tracker chan empty, fqdns chan string, gather chan []result,
	serverAddr string) {
	for fqdn := range fqdns {
		// fmt.Println(fqdn)
		results := DnsLookup(fqdn, serverAddr)
		if len(results) > 0 {
			gather <- results
		}
	}
	var e empty = empty{}
	tracker <- e
}

func DnsLookupA(fqdn, serverAddr string) ([]string, error) {
	var m dns.Msg
	var ips []string
	m.SetQuestion(dns.Fqdn(fqdn), dns.TypeA)
	in, err := dns.Exchange(&m, getDnsServerWithPort(serverAddr))
	if err != nil {
		return ips, err
	}

	if len(in.Answer) < 1 {
		return ips, errors.New("no answer")
	}
	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	return ips, nil
}

func DnsLookupCNAME(fqdn, serverAddr string) ([]string, error) {
	var m dns.Msg
	var fqdns []string
	m.SetQuestion(dns.Fqdn(fqdn), dns.TypeCNAME)
	in, err := dns.Exchange(&m, getDnsServerWithPort(serverAddr))
	if err != nil {
		return fqdns, err
	}
	if len(in.Answer) < 1 {
		return fqdns, errors.New("no answer")
	}
	for _, answer := range in.Answer {
		if c, ok := answer.(*dns.CNAME); ok {
			fqdns = append(fqdns, c.Target)
		}
	}
	return fqdns, nil
}

func DnsLookup(fqdn, serverAddr string) []result {
	var results []result
	var cfqdn = fqdn // Не изменять оригинал
	for {
		cnames, err := DnsLookupCNAME(cfqdn, serverAddr)
		if err == nil && len(cnames) > 0 {
			cfqdn = cnames[0]
			continue // Нужно обработать следующее CNAME
		}
		ips, err := DnsLookupA(cfqdn, serverAddr)
		// fmt.Println(cfqdn, ips)
		if err != nil {
			break // Для этого имени хоста А-записей нет
		}
		for _, ip := range ips {
			results = append(results, result{IPAddress: ip, Hostname: fqdn})
		}
		break // Все результаты были обработаны
	}
	return results
}

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve ",
	Short: "resolve dns to ip address",
	Long: `
resolve dns to ip address
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("resolve called with args:")
		// fmt.Println(args)
		var dnsServer, output string
		dnsServer, _ = cmd.Flags().GetString("dns-server")
		var isVarSet bool = false

		// isVarSet := cmd.Flags().Lookup("dns-server").Changed
		// if !isVarSet && len(Config.Dns.Common.DnsServer) > 0 {
		// 	dnsServer = Config.Dns.Common.DnsServer
		// }
		output, _ = cmd.Flags().GetString("output")
		isVarSet = cmd.Flags().Lookup("output").Changed
		if !isVarSet && len(Config.Common.OutputFormat) > 0 {
			output = Config.Common.OutputFormat
		}
		prettyJson, _ := cmd.Flags().GetBool("pretty-json")

		// resolveCmd.Flags().StringP("words-file", "w", "", "file with words to suggest existing subdomain")
		// resolveCmd.Flags().StringP("substitution-file", "z", "", "file with substitutions to resolve template")
		// resolveCmd.Flags().StringP("template", "t", "", "template to reolve fqdn names by substitutions from substitution-file")
		// resolveCmd.Flags().IntP("worker-number", "g", 3, "number of parallel workers")

		wordsFile, _ := cmd.Flags().GetString("words-file")
		substitutionFile, _ := cmd.Flags().GetString("substitution-file")
		template, _ := cmd.Flags().GetString("template")
		workerNumber, _ := cmd.Flags().GetInt("worker-number")

		_ = wordsFile
		_ = substitutionFile
		_ = template
		_ = workerNumber

		if len(args) == 0 && template == "" {
			Elogger.Err(errors.New("no domains or template provided")).Msg("no domains or templates provided in arguments or params")
			os.Exit(1)
		}
		var results []result
		fqdns := make(chan string, workerNumber)
		gather := make(chan []result)
		tracker := make(chan empty)

		// Start workers
		for i := 0; i < workerNumber; i++ {
			go DnsWorker(tracker, fqdns, gather, dnsServer)
		}

		var resultMap map[string]DnsResolveOutData = make(map[string]DnsResolveOutData)
		var resultPlain strings.Builder

		// results collector goroutine
		go func() {
			for r := range gather {
				results = append(results, r...)
			}
			var e empty
			tracker <- e
		}()

		// ❶ Send payload to workers

		// if just resolve by full domain names
		if len(args) > 0 && wordsFile == "" && template == "" {
			for _, domainName := range args {
				fqdns <- domainName
				resultMap[domainName] = DnsResolveOutData{HostName: domainName, IpAdresses: make([]net.IP, 0), DnsResolver: dnsServer}
			}
		}

		// if resolve subdomains by words-file and domain names
		if len(args) > 0 && wordsFile != "" && template == "" {
			var lines []string
			if fileContent, err := os.ReadFile(wordsFile); err != nil {
				Elogger.Err(err).Msg(err.Error())
				os.Exit(1)
			} else {
				lines = strings.Split(string(fileContent), "\n")
			}

			for _, domainName := range args {
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" && line != "." {
						words := strings.Split(line, ",")
						for _, word := range words {
							word = strings.TrimSpace(word)
							fqdns <- fmt.Sprintf("%s.%s", word, domainName)
						}
					}
					if line != "" && line == "." {
						fqdns <- domainName
					}
				}
				// resultMap[domainName] = DnsResolveOutData{HostName: domainName, IpAdresses: make([]net.IP, 0), DnsResolver: dnsServer}
			}
		}

		if len(args) == 0 && substitutionFile != "" && template != "" {
			// go run . dns resolve -t {1}.ya.{2} -s 192.168.89.1 -p -z /tmp/subst -o json
			// /tmp/subst :
			// {1}=main
			// {1}=api
			// {1}=stage
			// {2}=ru
			// {2}=com
			var lines []string
			if fileContent, err := os.ReadFile(substitutionFile); err != nil {
				Ilogger.Info().Msg("substitution file is not accesible - resolving template as is")
			} else {
				lines = strings.Split(string(fileContent), "\n")
			}

			subst := make(map[string][]string)
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && line != "." {
					rules := strings.Split(line, ",")
					for _, rule := range rules {
						leftAndRight := strings.Split(strings.TrimSpace(rule), "=")
						if len(leftAndRight) > 1 {
							left := strings.TrimSpace(leftAndRight[0])
							right := strings.TrimSpace(leftAndRight[1])
							subst[left] = append(subst[left], right)
						}
					}
				}
			}
			fqdnResults := GenerateSubstitutions(template, subst)
			for _, r := range fqdnResults {
				fqdns <- r
			}
		}

		close(fqdns)
		// Wait for workers to finish
		for i := 0; i < workerNumber; i++ {
			<-tracker
		}
		// close collector goroutine
		close(gather)
		// wait for collector goroutine to finish
		<-tracker

		// now we have results, we can print them
		for _, r := range results {
			ips := make([]net.IP, 0)
			existResultItem, ok := resultMap[r.Hostname]
			if !ok {
				existResultItem = DnsResolveOutData{HostName: r.Hostname, IpAdresses: ips, DnsResolver: dnsServer}
			}
			existResultItem.IpAdresses = append(existResultItem.IpAdresses, net.ParseIP(r.IPAddress))
			resultMap[r.Hostname] = existResultItem
		}

		if output == "json" {
			if prettyJson {
				stdOutResult, err := mcli_utils.PrettyJsonEncodeToString(resultMap)
				if err != nil {
					Elogger.Err(err).Msg(err.Error())
					os.Exit(1)
				}
				fmt.Println(stdOutResult)
			} else {
				fmt.Println(resultMap)
			}
		} else {
			for _, r := range resultMap {
				resultPlain.WriteString(fmt.Sprintf("%s\n", r.HostName))
				resultPlain.WriteString(fmt.Sprintf("%s\n", "--------------------------------"))
				if len(r.IpAdresses) > 0 {
					for _, ip := range r.IpAdresses {
						resultPlain.WriteString(fmt.Sprintf("%s\n", ip))
					}
				} else {
					resultPlain.WriteString(fmt.Sprintf("%s\n", "no ip addresses"))
				}
				resultPlain.WriteString(fmt.Sprintf("%s\n", ""))
			}
			fmt.Println()
			fmt.Println(resultPlain.String())
		}
	},
}

func init() {
	dnsCmd.AddCommand(resolveCmd)

	// Cobra supports local flags which will only run when this command is called directly, e.g.:
	resolveCmd.Flags().StringP("dns-server", "s", "8.8.8.8", "dns server - resolver")
	resolveCmd.Flags().StringP("output", "o", "plain", "output format: plain | json")
	resolveCmd.Flags().StringP("words-file", "w", "", "file with words to suggest existing subdomain")
	resolveCmd.Flags().StringP("substitution-file", "z", "", "file with substitutions to resolve template")
	resolveCmd.Flags().StringP("template", "t", "", "template to reolve fqdn names by substitutions from substitution-file")
	resolveCmd.Flags().IntP("worker-number", "g", 3, "number of parallel workers")
	resolveCmd.Flags().BoolP("pretty-json", "p", false, "pretty json output")

	// Cobra supports Persistent Flags which will work for this command and all subcommands, e.g.:
	// resolveCmd.PersistentFlags().String("foo", "", "A help for foo")

}

// Генерация всех возможных комбинаций подстановок
func GenerateSubstitutions(template string, subst map[string][]string) []string {
	// Регулярное выражение для поиска ключей из subst
	pattern := regexp.MustCompile(strings.Join(keysToRegex(subst), "|"))

	// Рекурсивная функция для генерации замен
	var generate func(string) []string
	generate = func(input string) []string {
		match := pattern.FindStringIndex(input)
		if match == nil {
			// Если больше нет совпадений, возвращаем строку как есть
			return []string{input}
		}

		// Извлекаем совпавший ключ
		start, end := match[0], match[1]
		key := input[start:end]

		// Массив для хранения всех комбинаций
		var results []string

		// Подставляем каждое значение из subst[key] и рекурсивно вызываем функцию
		for _, value := range subst[key] {
			replaced := input[:start] + value + input[end:]
			results = append(results, generate(replaced)...)
		}
		return results
	}

	// Начинаем генерацию
	return generate(template)
}

// Вспомогательная функция для преобразования ключей мапы в регулярное выражение
func keysToRegex(subst map[string][]string) []string {
	var keys []string
	for key := range subst {
		keys = append(keys, regexp.QuoteMeta(key)) // Экранируем спецсимволы
	}
	return keys
}

func oldResolve() {
	var msg dns.Msg
	var fqdn string
	var dnsServer string
	var resultPlain strings.Builder
	args := []string{}
	var resultMap map[string]DnsResolveOutData = make(map[string]DnsResolveOutData)
	if false {
		// secuantial processing
		for _, domainName := range args {
			fqdn = dns.Fqdn(domainName)
			msg.SetQuestion(fqdn, dns.TypeA)
			in, err := dns.Exchange(&msg, getDnsServerWithPort(dnsServer))
			if err != nil {
				Elogger.Err(err).Msg(err.Error())
				os.Exit(1)
			}
			resultDataItem := DnsResolveOutData{
				HostName:    fqdn,
				DnsResolver: dnsServer,
				IpAdresses:  []net.IP{},
			}
			resultPlain.WriteString(fqdn + "\n")
			resultPlain.WriteString("----------" + "\n")

			if len(in.Answer) < 1 {
				resultPlain.WriteString("No records" + "\n")
				resultPlain.WriteString("\n")
				continue
			}
			for _, answer := range in.Answer {
				if a, ok := answer.(*dns.A); ok {
					resultDataItem.IpAdresses = append(resultDataItem.IpAdresses, a.A)
					resultPlain.WriteString(a.A.String() + "\n")
				}
			}
			resultPlain.WriteString("\n")
			resultMap[fqdn] = resultDataItem
		}
	}
}
