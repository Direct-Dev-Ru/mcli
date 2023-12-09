package mclihttp

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/pkgerrors"
)

// CORSConfig represents the CORS configuration read from the file
type CORS struct {
	InfoLog        zerolog.Logger
	ErrorLog       zerolog.Logger
	corsFilepath   string
	AllowedDomains map[string]AllowedMethods `json:"allowed_domains"`
	Inner          http.Handler
}

// AllowedMethods represents the allowed methods for a specific domain
type AllowedMethods struct {
	Methods []string `json:"methods"`
}

func NewCORS(outILog zerolog.Logger, outErrLog zerolog.Logger, corsFilepath string) *CORS {
	cors := CORS{InfoLog: outILog, ErrorLog: outErrLog, corsFilepath: corsFilepath}

	err := cors.LoadCORSConfig(corsFilepath)
	if err != nil {
		outErrLog.Err(err).Msgf("%v", "create cors struct error ")
		return nil
	}

	return &cors
}

func (cors *CORS) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Handle preflight requests
	if req.Method == http.MethodOptions {
		cors.Inner.ServeHTTP(res, req)
	}
	origin := req.Header.Get("Origin")
	cors.InfoLog.Trace().Msgf("cors check. origin =: %v", origin)

	if _, ok := cors.AllowedDomains["*"]; ok || len(origin) == 0 || origin == "localhost" {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	} else {
		if methods, ok := cors.AllowedDomains[origin]; ok {
			// Set CORS headers for the allowed domain and methods
			res.Header().Set("Access-Control-Allow-Origin", origin)
			res.Header().Set("Access-Control-Allow-Methods", strings.Join(methods.Methods, ","))
			res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		} else {
			// Deny CORS for all other domains
			res.WriteHeader(http.StatusForbidden)
			res.Write([]byte("Forbidden - CORS not allowed for this domain"))
		}
	}
	cors.Inner.ServeHTTP(res, req)
}

func (cors *CORS) SetInnerHandler(next http.Handler) {
	cors.Inner = next
}

func (cors *CORS) LoadCORSConfig(filepath string) error {
	var reader io.Reader
	var err error
	if filepath == "" && cors.corsFilepath == "" {
		defaultCORS := `
		{
			"allowed_domains": {
				"*": {
					"methods": [
						"GET",
						"POST",
						"OPTIONS",
						"PUT",
						"DELETE"
					]
				}
			}
		}`
		reader = strings.NewReader(defaultCORS)
	} else {
		if filepath == "" && cors.corsFilepath != "" {
			filepath = cors.corsFilepath
		}
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer file.Close()
		reader = file
	}

	var config CORS
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	cors.AllowedDomains = config.AllowedDomains
	return nil
}
