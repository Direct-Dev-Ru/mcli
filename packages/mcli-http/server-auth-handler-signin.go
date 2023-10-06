package mclihttp

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	// mcli_utils "mcli/packages/mcli-utils"

	"net/http"
	"net/url"
	"strings"
	"time"
)

type signInDataMember struct {
	Action   string
	Redirect string
}
type signInData struct {
	Data signInDataMember
}

var tmplSignIn string = `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Sign in form</title>
			<link rel="stylesheet" href="/static/css/bootstrap.min.css">
		</head>
		<body>
			<div class="container mt-5">
				<div class="row justify-content-center">
					<div class="col-md-4">
						<h2 class="mb-4">Login</h2>
						<form method="POST" action="{{ .Data.Action }}">
							<div class="mb-3">
								<label for="username" class="form-label">Username</label>
								<input type="text" id="username" name="username" class="form-control" required>
							</div>
							<div class="mb-3">
								<label for="password" class="form-label">Password</label>
								<input type="password" id="password" name="password" class="form-control" required>
							</div>
							<button type="submit" class="btn btn-primary">Login</button>
						</form>
					</div>
				</div>
			</div>
		</body>
		</html>
	`

/*
var tmplSignIn string = `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Login Example</title>
		<style>
			.container {
				margin-top: 50px;
			}
			.form-container {
				max-width: 400px;
				margin: 0 auto;
				padding: 20px;
				border: 1px solid #ccc;
				border-radius: 5px;
			}
			.form-group {
				margin-bottom: 15px;
			}
			.form-label {
				display: block;
				margin-bottom: 5px;
			}
			.form-control {
				width: 90%;
				padding: 10px;
				border: 1px solid #ccc;
				border-radius: 3px;
			}
			.btn-primary {
				background-color: #007bff;
				color: #fff;
				padding: 10px 20px;
				border: none;
				border-radius: 3px;
				cursor: pointer;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="form-container">
				<h2>Login</h2>
				<form method="POST" action="/login">
					<div class="form-group">
						<label for="username" class="form-label">Username</label>
						<input type="text" id="username" name="username" class="form-control" required>
					</div>
					<div class="form-group">
						<label for="password" class="form-label">Password</label>
						<input type="password" id="password" name="password" class="form-control" required>
					</div>
					<button type="submit" class="btn-primary">Login</button>
				</form>
			</div>
		</div>
	</body>
	</html>
`
*/

var tmplParsed *template.Template
var loginData signInData

func GetSignInHandler(signInTemplatePath, baseUrl, action, redirect string) (HandleFunc, error) {

	tmplContent, err := os.ReadFile(signInTemplatePath)
	if err != nil {
		return nil, err
	}
	tmplSignIn = string(tmplContent)

	overAllActionUrl := strings.TrimPrefix(action, "/")
	if !strings.HasPrefix(action, baseUrl) {
		overAllActionUrl = fmt.Sprintf("%s/%s", baseUrl, overAllActionUrl)
	}
	overAllActionUrl = fmt.Sprintf("/%s", overAllActionUrl)

	loginData = signInData{Data: signInDataMember{Action: overAllActionUrl, Redirect: redirect}}
	tmplParsed, err = template.New("signin").Parse(tmplSignIn)
	if err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {
		signIn(w, r, tmplParsed, loginData)
	}, nil
}

func signIn(w http.ResponseWriter, r *http.Request, template *template.Template, loginData signInData) {
	if r.Method == http.MethodGet {
		tmplParsed.Execute(w, loginData)
		return
	}

	if r.Method == http.MethodPost {
		router, ok := r.Context().Value(ContextKey("router")).(*Router)
		if !ok {
			http.Error(w, "no router object in context", http.StatusInternalServerError)
			return
		}
		session := NewSession(cookieName, router.KVStore)
		session.Expire = time.Duration(timeExeed)

		var cred Credential = Credential{Expired: true, CredStore: router.CredentialStore}

		contentType := r.Header.Get("Content-Type")

		password, username := "", ""
		switch contentType {
		case "application/json":
			_, err := processJSONBody(w, r, &cred)

			if err != nil {
				http.Error(w, "wrong data in request body", http.StatusBadRequest)
				clearAuthenticatedCookie(w, session)
				return
			}
		case "application/x-www-form-urlencoded":
			formValues, _ := processFormValues(w, r)
			password, _ = getFormValue("password", formValues)
			username, _ = getFormValue("username", formValues)
			cred.SetCredential(username, password)
			// fmt.Println(username, password)
		default:
			http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
			clearAuthenticatedCookie(w, session)
			return
		}
		// fmt.Println("Cred getting from request: ", cred)

		if len(cred.Username) == 0 {
			http.Error(w, "username is empty in request body", http.StatusBadRequest)
			clearAuthenticatedCookie(w, session)
			return
		}

		_, err := session.SetToken("")
		if err != nil {
			http.Error(w, "auth get token error", http.StatusUnauthorized)
			clearAuthenticatedCookie(w, session)
			return
		}
		ok, err = session.Authenticate(cred)
		if !ok || err != nil {
			http.Error(w, "auth error: "+err.Error(), http.StatusUnauthorized)
			clearAuthenticatedCookie(w, session)
			return
		}

		// Finally, we set the client cookie for "session_token" as the session token we just generated
		// we also set an expiry time of 120 seconds, the same as the cache
		// http.SetCookie(w, &http.Cookie{
		// 	Name:    session.CookieName + "-plain",
		// 	Value:   session.Token,
		// 	Expires: time.Now().Add(session.Expire * time.Second),
		// })
		err = setAuthenticatedCookie(w, session)
		if err != nil {
			http.Error(w, "auth error: cookie setting error"+err.Error(), http.StatusUnauthorized)
			clearAuthenticatedCookie(w, session)
			return
		}

		if contentType == "application/json" {
			// Respond with JSON
			iUser, err, _ := cred.CredStore.GetUser(cred.Username)
			if err != nil {
				http.Error(w, "auth error: full user data get error: "+err.Error(), http.StatusUnauthorized)
				clearAuthenticatedCookie(w, session)
				return
			}
			fullUser, ok := iUser.(*Credential)
			if !ok {
				http.Error(w, "auth error: full user data get error: type mismatch "+err.Error(), http.StatusUnauthorized)
				clearAuthenticatedCookie(w, session)
				return
			}

			fullUser.Password = ""
			responseJSON := map[string]interface{}{
				"message": "Login successful",
				"error":   false,
				"payload": fullUser,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(responseJSON)
		} else {
			// Respond with plain text
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "Login successful")
		}

		return
	}
}

func clearAuthenticatedCookie(w http.ResponseWriter, session *Session) {
	cookie := &http.Cookie{
		Name:   session.CookieName,
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func processJSONBody(w http.ResponseWriter, r *http.Request, data interface{}) (interface{}, error) {
	// var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		// http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return nil, err
	}
	// fmt.Fprintln(w, "Received JSON data:", data)
	return data, nil
}

func processFormValues(w http.ResponseWriter, r *http.Request) (url.Values, error) {
	err := r.ParseForm()
	if err != nil {
		// http.Error(w, "Invalid form data", http.StatusBadRequest)
		return nil, err
	}

	formData := r.Form
	// Do something with the form data
	// fmt.Fprintln(w, "Received form values:", formData)
	return formData, nil
}

func getFormValue(fieldName string, values url.Values) (string, error) {
	if fieldValues, ok := values[fieldName]; ok {
		if len(fieldValues) == 1 {
			return fieldValues[0], nil
		}
		return strings.Join(fieldValues, "|"), nil
	}
	return "", fmt.Errorf("invalid field %s", fieldName)
}

func setAuthenticatedCookie(w http.ResponseWriter, session *Session) error {
	var cookieValueToStore = session.Token
	var err error
	if encodeCookies {
		cookieValueToStore, err = s.Encode(session.CookieName, session.Token)
	}
	if err == nil {
		cookie := &http.Cookie{
			Name:    session.CookieName,
			Value:   cookieValueToStore,
			Expires: time.Now().Add(session.Expire * time.Second),
		}
		http.SetCookie(w, cookie)
	}
	return err
}
