/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type MyStringKey string

// reverseCmd represents the reverse command
var reverseCmd = &cobra.Command{
	Use:   "reverse",
	Short: "Simple http(s) reverse proxy",
	Long: `Simple http(s) reverse proxy.
Example usage:
.....
`,
	Run: func(cmd *cobra.Command, args []string) {
		var maxIdleConns = 100
		var readTimeout = 30 * time.Second
		var writeTimeout = 30 * time.Second
		var idleConnTimeout = 60 * time.Second
		var baseURL = ""

		var host, port, baseUrl, tlsKey, tlsCert string
		host, _ = cmd.Flags().GetString("host")
		port, _ = cmd.Flags().GetString("port")
		baseUrl, _ = cmd.Flags().GetString("base-url")

		tlsKey, _ = cmd.Flags().GetString("tls-key")
		tlsCert, _ = cmd.Flags().GetString("tls-cert")

		intReadTimeout, _ := cmd.Flags().GetInt("read-timeout")
		readTimeout = time.Duration(intReadTimeout) * time.Second
		intWriteTimeout, _ := cmd.Flags().GetInt("write-timeout")
		writeTimeout = time.Duration(intWriteTimeout) * time.Second
		intIdleTimeout, _ := cmd.Flags().GetInt("idle-timeout")
		idleConnTimeout = time.Duration(intIdleTimeout) * time.Second

		// Parse the base URL
		parsedURL, err := url.Parse(baseUrl) // Replace with your target URL
		if err != nil {
			panic(err)
		}
		baseURL = parsedURL.String()

		// Configure client to handle timeouts and connections
		httpClient := &http.Client{
			Timeout: readTimeout + writeTimeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: maxIdleConns,
				DialContext: (&net.Dialer{
					Timeout:   readTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: readTimeout,
			},
		}

		// ReverseProxyHandler is the standard HTTP handler that proxies requests to the target server
		ReverseProxyHandler := func(w http.ResponseWriter, r *http.Request) {
			baseURL := baseURL
			// Extract target endpoint from query parameter
			targetEndpoint := r.URL.Query().Get("target")
			if targetEndpoint != "" {
				// Check if targetEndpoint is a full URL
				if _, err := url.ParseRequestURI(targetEndpoint); err != nil {
					// Handle as relative path
					if !strings.HasPrefix(targetEndpoint, "/") {
						targetEndpoint = "/" + targetEndpoint
					}
					parsedBaseUrl, _ := url.Parse(baseURL)
					hostBaseUrl := parsedBaseUrl.Host
					baseURL = parsedBaseUrl.Scheme + "://" + strings.TrimRight(hostBaseUrl, "/") + targetEndpoint
				} else {
					// It's a valid full URL, replace baseURL
					baseURL = targetEndpoint
				}
			}
			fmt.Printf("New baseURL is: %s\n", baseURL)

			ctx := context.WithValue(r.Context(), MyStringKey("baseURL"), baseURL)
			r = r.WithContext(ctx)
			// Start timing
			start := time.Now()

			proxyRequest, err := CreateProxyRequest(baseURL, targetEndpoint, httpClient, r)
			if err != nil {
				http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
				return
			}

			targetResp, err := httpClient.Do(proxyRequest)
			if err != nil {
				http.Error(w, "Failed to forward request", http.StatusBadGateway)
				return
			}
			defer targetResp.Body.Close()
			// Log the time taken to get the response
			duration := time.Since(start)
			log.Printf("Response received in: %v", duration)

			// Log the request in Apache style
			log.Printf("%s - - [%s] \"%s %s%s %s\" %d %d %v",
				r.RemoteAddr,
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				baseURL,
				r.URL.String(),
				r.Proto,
				targetResp.StatusCode,
				targetResp.ContentLength,
				duration,
			)

			CopyHeaders(w.Header(), targetResp.Header)
			w.WriteHeader(targetResp.StatusCode)
			io.Copy(w, targetResp.Body)
		}

		// Create a new HTTP server with the reverse proxy handler
		mux := http.NewServeMux()
		mux.HandleFunc("/", ReverseProxyHandler)
		// fmt.Println(tlsCert, tlsKey, host, port)
		var srv *http.Server
		if tlsCert != "" && tlsKey != "" {
			// Load your certificate and key files
			cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
			if err != nil {
				fmt.Println("Error loading TLS keys:", err)
				return
			}
			// Start the HTTPS server
			srv = &http.Server{
				Addr: fmt.Sprintf("%s:%s", host, port),
				TLSConfig: &tls.Config{
					Certificates:             []tls.Certificate{cert},
					MinVersion:               tls.VersionTLS13,
					PreferServerCipherSuites: true,
				},
				Handler:      mux,
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
				IdleTimeout:  idleConnTimeout,
			}
			fmt.Printf("Starting reverse HTTPS server to %s on %s\n", baseURL, srv.Addr)
			if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
				fmt.Println("Error starting HTTPS server:", err)
			}
		}

		if tlsCert == "" || tlsKey == "" {
			// Start the HTTP server
			srv = &http.Server{
				Addr:         fmt.Sprintf("%s:%s", host, port),
				Handler:      mux,
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
				IdleTimeout:  idleConnTimeout,
			}
			fmt.Printf("Starting reverse HTTP server to %s on %s\n", baseURL, srv.Addr)
			if err := srv.ListenAndServe(); err != nil {
				fmt.Println("Error starting HTTP server:", err)
			}
		}

	},
}

func init() {
	httpCmd.AddCommand(reverseCmd)

	reverseCmd.Flags().StringP("port", "p", "8080", "Specify port for reverse proxy server. default: 8080")
	reverseCmd.Flags().StringP("host", "H", "0.0.0.0", "Specify host for reverse proxy server. default: 0.0.0.0")
	reverseCmd.Flags().String("base-url", "", "Specify base url (proxied) path for reverse proxy server")
	reverseCmd.Flags().String("tls-cert", "", "Specify tls-cert file")
	reverseCmd.Flags().String("tls-key", "", "Specify tls-key file")
	reverseCmd.Flags().IntP("read-timeout", "", 30, "Specify read timeout")
	reverseCmd.Flags().IntP("write-timeout", "", 30, "Specify write timeout")
	reverseCmd.Flags().IntP("idle-timeout", "", 30, "Specify idle timeout")
}

func CreateProxyRequest(baseURL, targetEndpoint string, client *http.Client, r *http.Request) (*http.Request, error) {
	targetURL := baseURL + r.URL.String()
	if targetEndpoint != "" {
		targetURL = baseURL
	}
	proxyRequest, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		return nil, err
	}
	CopyHeaders(proxyRequest.Header, r.Header)

	return proxyRequest, nil
}

func CopyHeaders(dst http.Header, src http.Header) {
	for key := range src {
		dst[key] = src[key]
	}
}
