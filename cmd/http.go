/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_http "mcli/packages/mcli-http"
	mcli_redis "mcli/packages/mcli-redis"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts simple http server for testing purposes and it is top level command for other http servers and services",
	Long: `Starts simple http server for testing purposes and it is top level command for other http servers and services
	For example: mcli http --static-path ./http-static --static-prefix static
	it lookup index.html in ./http-static/html/index.html, other static assets in ./http-static/...
	and we can refer to it in url by /static/... prefix
`,
	Run: func(cmd *cobra.Command, args []string) {
		var port, baseUrl, staticPath, staticPrefix, tlsKey, tlsCert string
		var timeout int64 = 0

		port, _ = cmd.Flags().GetString("port")
		isPortSet := cmd.Flags().Lookup("port").Changed
		staticPath, _ = cmd.Flags().GetString("static-path")
		isStaticPathSet := cmd.Flags().Lookup("static-path").Changed
		staticPrefix, _ = cmd.Flags().GetString("static-prefix")
		isStaticPrefix := cmd.Flags().Lookup("static-prefix").Changed
		baseUrl, _ = cmd.Flags().GetString("base-url")
		isBaseUrl := cmd.Flags().Lookup("base-url").Changed

		tlsKey, _ = cmd.Flags().GetString("tls-key")
		tlsCert, _ = cmd.Flags().GetString("tls-cert")
		timeout, _ = cmd.Flags().GetInt64("timeout")
		isTimeoutSet := cmd.Flags().Lookup("timeout").Changed

		// process configuration or setup defaults
		if !isPortSet && len(Config.Http.Server.Port) > 0 {
			port = Config.Http.Server.Port
		}
		if !isTimeoutSet && Config.Http.Server.Timeout > 0 {
			timeout = Config.Http.Server.Timeout
		}
		if !isStaticPathSet && len(Config.Http.Server.StaticPath) > 0 {
			staticPath = Config.Http.Server.StaticPath
		}
		if !isStaticPrefix && len(Config.Http.Server.StaticPrefix) > 0 {
			staticPrefix = Config.Http.Server.StaticPrefix
		}
		if !isBaseUrl && len(Config.Http.Server.BaseUrl) > 0 {
			baseUrl = Config.Http.Server.BaseUrl
		}

		tmplPath, _ := GetStringParam("tmpl-path", cmd, Config.Http.Server.TmplPath)
		tmplPrefix, _ := GetStringParam("tmpl-prefix", cmd, Config.Http.Server.TmplPrefix)
		tmplDataPath, _ := GetStringParam("tmpl-datapath", cmd, Config.Http.Server.TmplDataPath)

		// Channel for interrupt signal
		StopHttpChan := make(chan os.Signal, 1)

		// mcli_http.InitMainRoutes(staticPath, staticPrefix)
		r := mcli_http.NewRouter(staticPath, staticPrefix, Ilogger, Elogger, mcli_http.RouterOptions{BaseUrl: baseUrl})

		// root route
		rootRoute := mcli_http.NewRoute("/", mcli_http.Equal)
		rootRoute.SetHandler(func(res http.ResponseWriter, req *http.Request) {

			// authUser, ok := req.Context().Value(mcli_http.ContextKey("authUser")).(string)
			// fmt.Println("authUser store in ctx", ok, authUser)

			sPath := "./" + staticPath
			if strings.HasPrefix(staticPath, "/") {
				sPath = staticPath
			}
			mainPagePath := ""
			mainPagePathCandidate := sPath + "/index.html"
			if _, err := os.Stat(mainPagePathCandidate); err != nil {
				mainPagePathCandidate = sPath + "/html/index.html"
				if _, err := os.Stat(mainPagePathCandidate); err != nil {
					mainPagePathCandidate = ""
				}
			}
			mainPagePath = mainPagePathCandidate

			if len(mainPagePath) > 0 {
				http.ServeFile(res, req, mainPagePath)
			} else {
				http.Error(res, "404 Not Found index.html", 404)
			}
		})
		r.AddRoute(rootRoute)
		// route for template handling
		// r.SetTmplRoutes(tmplPath, tmplPrefix, tmplDataPath)

		// Context to stop server and pass into other goroutines
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			// extra handling here
			//fmt.Println("extra handling done")
			cancel()
			time.Sleep(5 * time.Second)
		}()
		serverTemplates := Config.Http.Server.Templates

		if len(tmplPath) > 0 {
			serverTemplates = make([]mcli_http.TemplateEntry, 0, 1)
			serverTemplates[0] = mcli_http.TemplateEntry{TmplName: "fromcmdline", TmplType: "standart",
				TmplPath: tmplPath, TmplPrefix: tmplPrefix, TmplDataPath: tmplDataPath}
		}
		r.SetTemplatesRoutes(ctx, serverTemplates)

		r.AddRouteWithHandler("/echo", mcli_http.Prefix, mcli_http.Http_Echo)

		// setting up middleware
		// err := r.Use(mcli_http.NewLogger(Ilogger, Elogger, mcli_http.LoggerOpts{ShowUrl: true, ShowIp: false}))
		// if err != nil {
		// 	Elogger.Error().Err(err)
		// }

		if Config.Http.Server.Auth.IsAuthenticate {
			// _, err = mcli_redis.InitCache(Config.Http.Server.Auth.RedisHost, Config.Http.Server.Auth.RedisPwd)
			redisStore, err := mcli_redis.NewRedisStore(Config.Http.Server.Auth.RedisHost, Config.Http.Server.Auth.RedisPwd, "userlist")
			if err != nil {
				Elogger.Fatal().Err(fmt.Errorf("error init redis store: %v", err))
			}
			_, err = redisStore.RedisPool.Get().Do("PING")
			if err != nil {
				Elogger.Fatal().Msgf("redis connection error: %v", err.Error())
			}
			Ilogger.Trace().Msg("Ping Pong to redis server is successful")

			r.KVStore = redisStore
			r.CredentialStore = mcli_http.NewUserStore(redisStore, "userlist")

			internalSecretStorePath := filepath.Join(GlobalMap["RootPath"], "internal-secrets")
			_, _, err = mcli_utils.IsExistsAndCreate(internalSecretStorePath, true)
			if err != nil {
				Elogger.Fatal().Msgf("internal secret store error - path do not exists: %v", err.Error())
			}

			internalSecretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile, mcli_crypto.AesCypher, nil)
			internalVaultPath := filepath.Join(internalSecretStorePath, "intstore.vault")
			if err := internalSecretStore.FillStore(internalVaultPath, GlobalMap["RootSecretKeyPath"]); err != nil {
				Elogger.Fatal().Msg(err.Error())
			}
			secretMapa := internalSecretStore.GetSecretPlainMap()
			cookieKey1, cookieKey2 := "", ""
			cookieKey1Secret, ok := secretMapa["CookieKey1"]

			if ok {
				cookieKey1 = cookieKey1Secret.Secret
				if err != nil {
					Elogger.Fatal().Msg(err.Error())
				}
				Ilogger.Trace().Msg(cookieKey1)
			} else {
				cookieKey1 = string(mcli_secrets.GenKey(32))

				secretEntry1, err := internalSecretStore.NewEntry("CookieKey1", "CookieKey1", "Key 1 for Cookie encription")
				if err != nil {
					Elogger.Fatal().Msgf("cookieKey1 new entry error: %v", err)
				}
				secretEntry1.SetSecret(fmt.Sprintf("%x", cookieKey1), true, false)
				Ilogger.Trace().Msg(fmt.Sprintf("%x", cookieKey1))

				internalSecretStore.AddEntry(secretEntry1)
				internalSecretStore.Save(internalVaultPath, GlobalMap["RootSecretKeyPath"])
			}

			cookieKey2Secret, ok := secretMapa["CookieKey2"]

			if ok {
				cookieKey2 = cookieKey2Secret.Secret
				if err != nil {
					Elogger.Fatal().Msg(err.Error())
				}
				Ilogger.Trace().Msg(cookieKey2)
			} else {
				// time.Sleep(200 * time.Millisecond)
				cookieKey2 = string(mcli_secrets.GenKey(32))
				secretEntry2, err := internalSecretStore.NewEntry("CookieKey2", "CookieKey2", "Key 2 for Cookie encription")
				if err != nil {
					Elogger.Fatal().Msgf("cookieKey2 new entry error: %v", err)
				}
				secretEntry2.SetSecret(fmt.Sprintf("%x", cookieKey2), true, false)
				// Ilogger.Trace().Msg(fmt.Sprintf("%x", cookieKey2))
				internalSecretStore.AddEntry(secretEntry2)

				internalSecretStore.Save(internalVaultPath, GlobalMap["RootSecretKeyPath"])
			}
			isEncCookie := false
			if len(cookieKey1) > 0 && len(cookieKey2) > 0 {
				isEncCookie = true
				cookieKey1, err = mcli_secrets.LoadKeyFromHexString(cookieKey1)
				if err != nil {
					isEncCookie = false
				}
				cookieKey2, err = mcli_secrets.LoadKeyFromHexString(cookieKey2)
				if err != nil {
					isEncCookie = false
				}
			}
			mcli_http.SetSecretCookieOptions(isEncCookie, "session-token", []byte(cookieKey1), []byte(cookieKey2))

			// auth middleware
			err = r.Use(mcli_http.NewAuth(r.CredentialStore, r.KVStore, isEncCookie))
			if err != nil {
				Elogger.Error().Err(err)
			}

			r.AddRouteWithHandler(Config.Http.Server.Auth.SignInRoute, mcli_http.Prefix, mcli_http.Signin)

			// store := mcli_http.UserRedisStore{RedisPool: mcli_http.RedisPool, CollectionPrefix: "userlist"}
			// c, _ := store.GetAllUsers("")
			// for _, user := range c {
			// 	fmt.Println(*user)
			// }

		} else {
			Ilogger.Warn().Msg("Authentication and sessions are disabled !!!")
		}

		var srv *http.Server
		if len(tlsCert) > 0 && len(tlsKey) > 0 {

			srv = &http.Server{
				Addr:         ":" + port,
				Handler:      r,
				ReadTimeout:  time.Duration(timeout * int64(time.Millisecond)),
				WriteTimeout: time.Duration(timeout * 3 * int64(time.Millisecond)),
				IdleTimeout:  time.Duration(timeout * 4 * int64(time.Millisecond)),
				TLSConfig: &tls.Config{
					MinVersion:               tls.VersionTLS13,
					PreferServerCipherSuites: true,
				},
			}

			go func() {
				if Config.Http.Server.Auth.IsAuthenticate {
					defer mcli_redis.RedisPool.Close()
				}
				if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
					Elogger.Fatal().Msg(err.Error())
				}
			}()
		} else {
			srv = &http.Server{
				Addr:         ":" + port,
				Handler:      r,
				ReadTimeout:  time.Duration(timeout * int64(time.Millisecond)),
				WriteTimeout: time.Duration(timeout * 3 * int64(time.Millisecond)),
				IdleTimeout:  time.Duration(timeout * 4 * int64(time.Millisecond)),
			}

			go func() {
				if Config.Http.Server.Auth.IsAuthenticate {
					defer mcli_redis.RedisPool.Close()
				}
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					Elogger.Fatal().Msg(err.Error())
				}
			}()
		}

		signal.Notify(StopHttpChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		Ilogger.Info().Msg(fmt.Sprintf("http server started on port %s", port))

		<-StopHttpChan

		Ilogger.Info().Msg(fmt.Sprintf("http server stopped on port %s", port))

		if err := srv.Shutdown(ctx); err != nil {
			Elogger.Fatal().Msg(fmt.Sprintf("server shutdown failed: %+v", err))
		}
		Ilogger.Info().Msg("server shutting down properly")
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().Int64P("timeout", "t", 5000, "Specify timeout for http server service")

	// setup flags
	var port, staticPath, staticPrefix string = "8080", "http-static", "static"
	var tmplPath, tmplPrefix, tmplDataPath string = "", "", ""

	httpCmd.Flags().StringP("port", "p", port, "Specify port for test http server.")
	httpCmd.Flags().String("base-url", "", "Specify base url path for http server")
	httpCmd.Flags().String("static-path", staticPath, "Specify relative path to static folder")
	httpCmd.Flags().String("static-prefix", staticPrefix, "Specify url prefix part to static content")
	httpCmd.Flags().String("tmpl-path", tmplPath, "Specify relative or absolute path to template folder")
	httpCmd.Flags().String("tmpl-prefix", tmplPrefix, "Specify url prefix part to handle template content")
	httpCmd.Flags().String("tmpl-datapath", tmplDataPath, "Specify relative or absolute path to bson or json files for templates")
	httpCmd.Flags().String("tls-cert", "", "Specify tls-cert file")
	httpCmd.Flags().String("tls-key", "", "Specify tls-key file")
}
