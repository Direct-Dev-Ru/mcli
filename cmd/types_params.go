package cmd

type ConfigData struct {
	Common struct {
		OutputFile   string `yaml:"output-file"`
		OutputFormat string `yaml:"output-format"`
	}

	Http struct {
		Server struct {
			Timeout      int64  `yaml:"timeout"`
			Port         string `yaml:"port"`
			StaticPath   string `yaml:"static-path"`
			StaticPrefix string `yaml:"static-prefix"`
		}

		Request struct {
			Timeout int64                  `yaml:"timeout"`
			Method  string                 `yaml:"method"`
			BaseURL string                 `yaml:"baseURL"`
			URL     string                 `yaml:"url"`
			Headers map[string][]string    `yaml:"headers"`
			Body    map[string]interface{} `yaml:"body"`
		}
	}

	Secrets struct {
		Common struct {
			VaultPath   string `yaml:"vault-path"`
			KeyFilePath string `yaml:"keyfile-path"`
			DictPath    string `yaml:"dict-path"`
			UseWords    bool   `yaml:"use-words"`
			MinLength   int    `yaml:"min-lenght"`
			MaxLength   int    `yaml:"max-lenght"`
		}

		Export struct {
			ExportPath string `yaml:"export-path"`
		}
	}
}