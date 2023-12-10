package mclihttp

import (
	"fmt"
	"strings"
)

var HttpConfig Http

type Server struct {
	Timeout           int64           `yaml:"timeout"`
	Port              string          `yaml:"port"`
	BaseUrl           string          `yaml:"base-url"`
	StaticPath        string          `yaml:"static-path"`
	StaticPrefix      string          `yaml:"static-prefix"`
	TmplPath          string          `yaml:"tmpl-path"`
	TmplPrefix        string          `yaml:"tmpl-prefix"`
	TmplDataPath      string          `yaml:"tmpl-datapath"`
	Templates         []TemplateEntry `yaml:"templates"`
	CorsParamFilePath string          `yaml:"cors-filepath"`
	RouterV2          bool

	RootPage struct {
		RootPageTemplate     string `yaml:"rootpage-template"`
		RootPageTitle        string `yaml:"rootpage-title"`
		RedirectUnauthorized bool   `yaml:"redirect-unauthorized"`
	} `yaml:"root-page"`

	Auth struct {
		IsAuthenticate bool `yaml:"is-authenticate"`

		SignInRoute          string `yaml:"signin-route"`
		SignInTemplate       string `yaml:"signin-template"`
		SignInChangeRoute    string `yaml:"signin-change-route"`
		SignInChangeTemplate string `yaml:"signin-change-template"`
		SignInRedirect       string `yaml:"signin-redirect"`

		SignUpRoute           string `yaml:"signup-route"`
		SignUpTemplate        string `yaml:"signup-template"`
		SignUpConfirmRoute    string `yaml:"signup-confirm-route"`
		SignInConfirmTemplate string `yaml:"signup-confirm-template"`
		SignUpRedirect        string `yaml:"signup-redirect"`

		ProfileRoute    string `yaml:"profile-route"`
		ProfileTemplate string `yaml:"profile-template"`

		AuthTtl             int    `yaml:"auth-ttl"`
		SecureAuthToken     bool   `yaml:"secure-auth-token"`
		AuthTokenName       string `yaml:"auth-token-name"`
		SessionsRedisPrefix string `yaml:"sessions-redis-prefix"`

		RedisHost string `yaml:"redis-host"`
		RedisPwd  string `yaml:"redis-password"`
	} `yaml:"auth"`
}

type Request struct {
	Timeout int64                  `yaml:"timeout"`
	Method  string                 `yaml:"method"`
	BaseURL string                 `yaml:"base-url"`
	URL     string                 `yaml:"url"`
	Headers map[string][]string    `yaml:"headers"`
	Body    map[string]interface{} `yaml:"body"`
}

type Http struct {
	Server  Server
	Request Request
}

func (config *Http) GetFullUrl(url string, baseUrl ...string) string {
	base := config.Server.BaseUrl
	if len(baseUrl) > 0 {
		base = baseUrl[0]
	}
	overallUrl := strings.TrimPrefix(url, "/")
	if !strings.HasPrefix(overallUrl, base) {
		overallUrl = fmt.Sprintf("%s/%s", base, overallUrl)
	}
	// fmt.Println(base, url, baseUrl)
	// fmt.Println(fmt.Sprintf("/%s", overallUrl))
	return fmt.Sprintf("/%s", overallUrl)
}
