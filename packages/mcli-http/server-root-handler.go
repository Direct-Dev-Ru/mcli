package mclihttp

import (
	"fmt"
	"html/template"
	"os"

	// mcli_utils "mcli/packages/mcli-utils"

	"net/http"
	"strings"
)

type rootInDataMember struct {
	Action   string
	Title    string
	Redirect string
}
type rootInData struct {
	Data rootInDataMember
}

func GetRootHandler(signInTemplatePath, baseUrl, title, action, redirect string) (HandleFunc, error) {

	tmplRootDefault := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Site about ...</title>
		<link rel="stylesheet" href="/static/css/bootstrap.min.css">
	</head>
	<body>
		<div class="container mt-5">
			<div class="row justify-content-center">
				<div class="col-md-4">
					<h2 class="mb-4">Hello Gopher!!!</h2>
					<form method="POST" action="{{ .Data.Action }}">							
						<button type="submit" class="btn btn-primary">Start Journey</button>
					</form>
				</div>
			</div>
		</div>
	</body>
	</html>
	`

	var tmplRootParsed *template.Template
	var rootData rootInData

	tmplContent, err := os.ReadFile(signInTemplatePath)
	if err != nil {
		tmplContent = []byte(tmplRootDefault)
	}
	tmplRoot := string(tmplContent)

	overAllActionUrl := strings.TrimPrefix(action, "/")
	if !strings.HasPrefix(action, baseUrl) {
		overAllActionUrl = fmt.Sprintf("%s/%s", baseUrl, overAllActionUrl)
	}
	overAllActionUrl = fmt.Sprintf("/%s", overAllActionUrl)

	tmplRootParsed, err = template.New("root").Parse(tmplRoot)
	if err != nil {
		return nil, err
	}

	rootData = rootInData{Data: rootInDataMember{Action: overAllActionUrl, Title: title, Redirect: redirect}}

	return func(w http.ResponseWriter, r *http.Request) {
		rootHandler(w, r, tmplRootParsed, rootData)
	}, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request, template *template.Template, inData rootInData) {
	if r.Method == http.MethodGet {
		template.Execute(w, inData)
		return
	}
	// TODO: process POST handler
	http.Error(w, "method not supported", http.StatusNotFound)
}
