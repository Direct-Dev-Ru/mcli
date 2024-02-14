package cmd

import (
	mcli_http "mcli/packages/mcli-http"
	mcli_type "mcli/packages/mcli-type"
)

type ConfigData struct {
	Cache mcli_type.Cacher

	ConfigVersion string `yaml:"config-version"`
	Common        struct {
		AppName             string `yaml:"app-name"`
		OutputFile          string `yaml:"output-file"`
		OutputFormat        string `yaml:"output-format"`
		InternalKeyFilePath string `yaml:"internal-keyfile-path"`
		InternalVaultPath   string `yaml:"internal-vault-path"`
		RedisHost           string `yaml:"redis-host"`
		RedisPwd            string `yaml:"redis-password"`
		RedisDatabaseNo     int    `yaml:"redis-database-no"`
		RedisRequire        bool   `yaml:"redis-require"`
	}
	Http mcli_http.Http
	// Http struct {
	// 	Server struct {
	// 		Timeout      int64                     `yaml:"timeout"`
	// 		Port         string                    `yaml:"port"`
	// 		BaseUrl      string                    `yaml:"base-url"`
	// 		StaticPath   string                    `yaml:"static-path"`
	// 		StaticPrefix string                    `yaml:"static-prefix"`
	// 		TmplPath     string                    `yaml:"tmpl-path"`
	// 		TmplPrefix   string                    `yaml:"tmpl-prefix"`
	// 		TmplDataPath string                    `yaml:"tmpl-datapath"`
	// 		Templates    []mcli_http.TemplateEntry `yaml:"templates"`

	// 		RootPage struct {
	// 			RootPageTemplate     string `yaml:"rootpage-template"`
	// 			RootPageTitle        string `yaml:"rootpage-title"`
	// 			RedirectUnauthorized bool   `yaml:"redirect-unauthorized"`
	// 		} `yaml:"root-page"`

	// 		Auth struct {
	// 			IsAuthenticate bool   `yaml:"is-authenticate"`
	// 			SignInRoute    string `yaml:"signin-route"`
	// 			SignInTemplate string `yaml:"signin-template"`
	// 			SignInRedirect string `yaml:"signin-redirect"`
	// 			SignUpRoute    string `yaml:"signup-route"`
	// 			SignUpTemplate string `yaml:"signup-template"`
	// 			SignUpRedirect string `yaml:"signup-redirect"`
	// 			RedisHost      string `yaml:"redis-host"`
	// 			RedisPwd       string `yaml:"redis-password"`
	// 		} `yaml:"auth"`
	// 	}

	// 	Request struct {
	// 		Timeout int64                  `yaml:"timeout"`
	// 		Method  string                 `yaml:"method"`
	// 		BaseURL string                 `yaml:"base-url"`
	// 		URL     string                 `yaml:"url"`
	// 		Headers map[string][]string    `yaml:"headers"`
	// 		Body    map[string]interface{} `yaml:"body"`
	// 	}
	// }

	Secrets struct {
		Common struct {
			VaultPath   string `yaml:"vault-path"`
			KeyFilePath string `yaml:"keyfile-path"`
			KeyEnvVar   string `yaml:"key-envvar"`
			DictPath    string `yaml:"dict-path"`
			UseWords    bool   `yaml:"use-words"`
			Obfuscate   bool   `yaml:"obfuscate"`
			MinLength   int    `yaml:"min-lenght"`
			MaxLength   int    `yaml:"max-lenght"`
		}

		Export struct {
			ExportPath string `yaml:"export-path"`
		}
	}
}
