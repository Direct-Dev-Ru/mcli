/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	UrlPackage "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// func httpRequestReadBody(rc io.ReadCloser) (string, error) {
// 	data, err := io.ReadAll(rc)
// 	if err == nil {
// 		defer rc.Close()
// 		strBody, err := UrlPackage.QueryUnescape(string(data))
// 		if err == nil {
// 			return strBody, nil
// 		}
// 	}
// 	return "", err
// }

type httpRequestOpts struct {
	timeout             int64
	MaxIdleConns        int
	MaxConnsPerHost     int
	MaxIdleConnsPerHost int
	body                interface{}
	headers             map[string][]string
}

func httpRequestDo(ctx context.Context, method, url string, opts *httpRequestOpts) (*http.Response, error) {
	var response *http.Response
	var req *http.Request
	var err error
	if ctx == nil {
		ctx = context.Background()
	}
	mapHeaders := opts.headers

	body := opts.body
	var builder strings.Builder
	err = json.NewEncoder(&builder).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, method, url, io.NopCloser(strings.NewReader(builder.String())))

	if err == nil {
		req.Header["User-Agent"] = []string{"mcli v." + Version}
		for k, v := range mapHeaders {
			req.Header[k] = v
		}

		// do request
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = opts.MaxIdleConns
		t.MaxConnsPerHost = opts.MaxConnsPerHost
		t.MaxIdleConnsPerHost = opts.MaxIdleConnsPerHost

		client := &http.Client{
			Timeout:   time.Duration(opts.timeout * int64(time.Millisecond)),
			Transport: t,
		}
		response, err = client.Do(req)
		return response, err

	} else {
		return nil, err
	}
}

func loadFromJYSAOMNL(input, format string, isFile bool) (parsed interface{}, err error) {
	var in io.Reader
	if format == "" {
		format = "json"
	}
	var file *os.File
	if isFile {
		file, err = os.Open(input)
		if err != nil {
			parsed = nil
			return
		}
		defer file.Close()
		in = file
	} else {
		in = strings.NewReader(input)
	}
	parsed = struct{}{}
	if format == "yaml" {
		decodeYAML := yaml.NewDecoder(in)
		err = decodeYAML.Decode(&parsed)
	} else {
		decodeJSON := json.NewDecoder(in)
		err = decodeJSON.Decode(&parsed)
	}
	if err != nil {
		return
	}
	return
}

func headersParser(headers string) (headersData map[string][]string, err error) {
	ext := filepath.Ext(strings.ToLower(headers))
	isFile := ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".form"

	headersData = make(map[string][]string)
	var hs interface{}

	if isFile && (ext == ".yaml" || ext == ".yml") {
		hs, err = loadFromJYSAOMNL(headers, "yaml", isFile)
	}
	if isFile && ext == ".form" {
		hs, err = loadFromJYSAOMNL(headers, "form", isFile)
	}
	if isFile && ext == ".json" || !isFile {
		hs, err = loadFromJYSAOMNL(headers, "json", isFile)
	}

	headersMap, ok := hs.(map[string]interface{})
	if !ok {
		err = errors.New("wrong format for headers - should be map[string][]interface{}")
		return
	}
	for k, v := range headersMap {
		headerArray, ok := v.([]interface{})
		if !ok {
			err = errors.New("wrong format for headers - should be map[string][]interface{}")
			headersData = make(map[string][]string)
			return
		}
		s := make([]string, len(headerArray))
		for i, vv := range headerArray {
			s[i] = fmt.Sprint(vv)
		}
		headersData[k] = s
	}
	return
}

func bodyParser(body string) (bodyData map[string]interface{}, err error) {

	ext := filepath.Ext(strings.ToLower(body))

	isFile := ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".form"

	var bs interface{}

	if isFile && (ext == ".yaml" || ext == ".yml") {
		bs, err = loadFromJYSAOMNL(body, "yaml", isFile)
	}
	if isFile && ext == ".form" {
		bs, err = loadFromJYSAOMNL(body, "form", isFile)
	}
	if isFile && ext == ".json" || !isFile {
		bs, err = loadFromJYSAOMNL(body, "json", isFile)
	}

	bodyData, ok := bs.(map[string]interface{})
	if !ok {
		err = errors.New("wrong format for body - should be map[string]interface{}")
		return
	}

	return
}

// requestCmd represents the request command
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Do http request",
	Long: `With request command it is possible to do http requests. 
	config params in config file may be more priority then commandline params, but "headers" and "body" params concatenates
	For example: 
	mcli http request --url http://localhost:8080/echo -m POST -s '{"Content-Type":["application/json"]}'  \ 
		-b '{"id": 1001}'

`,
	Run: func(cmd *cobra.Command, args []string) {

		var method, baseURL, url, headers, body string
		var timeout int64 = 0
		var mapHeaders map[string][]string
		var mapBody map[string]interface{}

		timeout, _ = cmd.Flags().GetInt64("timeout")
		isTimeoutSet := cmd.Flags().Lookup("timeout").Changed

		method, _ = cmd.Flags().GetString("method")
		isMethodSet := cmd.Flags().Lookup("method").Changed

		url, _ = cmd.Flags().GetString("url")
		url = strings.TrimSpace(url)
		isUrlSet := cmd.Flags().Lookup("url").Changed

		headers, _ = cmd.Flags().GetString("headers")
		isHeadersSet := cmd.Flags().Lookup("headers").Changed

		body, _ = cmd.Flags().GetString("body")
		isBodySet := cmd.Flags().Lookup("body").Changed

		// process configuration or setup defaults
		if !isMethodSet && len(Config.Http.Request.Method) > 0 {
			method = Config.Http.Request.Method
		}
		if !isTimeoutSet && Config.Http.Request.Timeout > 0 {
			timeout = Config.Http.Request.Timeout
		}
		if !isUrlSet && len(Config.Http.Request.URL) > 0 {
			url = strings.TrimSpace(Config.Http.Request.URL)
		}
		if len(Config.Http.Request.BaseURL) > 0 {
			baseURL = strings.TrimSpace(Config.Http.Request.BaseURL)
		}

		if !isBodySet && len(Config.Http.Request.Body) > 0 {
			mapBody = Config.Http.Request.Body
		} else {
			var bErr error
			mapBody, bErr = bodyParser(body)
			if bErr != nil {
				Elogger.Error().Msg(fmt.Sprintf("body parsing error: %v ", bErr.Error()))
			}
		}

		if !isHeadersSet && len(Config.Http.Request.Headers) > 0 {
			mapHeaders = Config.Http.Request.Headers
		} else {
			var hErr error
			mapHeaders, hErr = headersParser(headers)
			if hErr != nil {
				Elogger.Error().Msg(fmt.Sprintf("headers parsing error: %v ", hErr.Error()))
			}
		}

		URL, err := UrlPackage.Parse(baseURL + url)
		if err != nil {
			Elogger.Fatal().Msg(fmt.Sprintf("fatal error while parsing url: %v ", err.Error()))
		}
		// fmt.Println(URL)
		// URL.Path = UrlPackage.PathEscape(URL.Path)
		// URL.RawQuery = UrlPackage.QueryEscape(URL.RawQuery)

		// Do http request
		// Ilogger.Trace().Msg(fmt.Sprintf("method: %v. url: %v %v", method, url, URL.String()))
		reqOpts := &httpRequestOpts{
			timeout:             timeout,
			body:                mapBody,
			headers:             mapHeaders,
			MaxIdleConns:        100,
			MaxConnsPerHost:     100,
			MaxIdleConnsPerHost: 100,
		}

		response, err := httpRequestDo(context.Background(), method, URL.String(), reqOpts)

		if response == nil {
			Elogger.Fatal().Msg(fmt.Sprintf("error: response is nil %v", err.Error()))
		}

		if err != nil {
			Elogger.Fatal().Msg(fmt.Sprintf("error: %v. status code: %v", err.Error(), response.Status))
		}

		defer response.Body.Close()

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			buf := new(strings.Builder)
			strBody := ""

			_, err := io.Copy(buf, response.Body)
			if err != nil {
				Elogger.Fatal().Msg(fmt.Sprintf("error: %v", err.Error()))
			}
			strBody, err = UrlPackage.QueryUnescape(buf.String())
			if err != nil {
				Elogger.Fatal().Msg(fmt.Sprintf("error: %v", err.Error()))
			}
			Ilogger.Trace().Msg(fmt.Sprintf("Response headers:\n %v", response.Header))
			Ilogger.Trace().Msg("Response body:\n" + strBody)
		} else {
			Elogger.Error().Msg(fmt.Sprintf("Response status code is %v", response.StatusCode))
		}

	},
}

func init() {
	httpCmd.AddCommand(requestCmd)

	var method, url string = "GET", ""
	requestCmd.Flags().StringP("method", "m", method, "Specify method for http request")
	requestCmd.Flags().StringP("url", "u", url, "Specify URL for http request")
	requestCmd.Flags().StringP("headers", "s", "", "Specify headers: {h1:[v1],h2:[v2,v3]}")
	requestCmd.Flags().StringP("body", "b", "", "Specify json representation of a body")
	requestCmd.Flags().StringP("form", "f", "", "Specify json representation of a form")
	requestCmd.Flags().Int64P("timeout", "t", 5000, "Specify timeout for http services (server and request)")
}
