package mclihttp

import (
	"errors"
	"fmt"
	mcli_interface "mcli/packages/mcli-interface"
	mcli_utils "mcli/packages/mcli-utils"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	sc "github.com/gorilla/securecookie"
)

var cookieName string = "session-token"
var (
	timeExeed     = 36000
	s             *sc.SecureCookie
	encodeCookies = false
)

func SetSecretCookieOptions(doEncoding bool, cookieName string, cookieHash, cookieBlock []byte) {
	if cookieName == "" {
		cookieName = "session-token"
	}
	encodeCookies = doEncoding
	s = nil
	if encodeCookies {
		s = sc.New(cookieHash, cookieBlock)
	}
}

type Credential struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Expired     bool     `json:"expired,string,omitempty"`
	FirstName   string   `json:"first-name"`
	LastName    string   `json:"last-name"`
	Email       string   `json:"email"`
	BackupEmail string   `json:"backup-email"`
	Phone       string   `json:"phone"`
	BackupPhone string   `json:"backup-phone"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
	CredStore   mcli_interface.CredentialStorer
}

// func (f *Credential) UnmarshalJSON(b []byte) (err error) {
// 	switch str := strings.ToLower(strings.Trim(string(b), `"`)); str {
// 	case "true":
// 		f.Expired = true
// 	case "false":
// 		f.Expired = false
// 	}
// 	return err
// }

func NewCredential(username, password string, expired bool, credStore mcli_interface.CredentialStorer) *Credential {
	username = strings.TrimSpace(username)
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	emailRegex, err := regexp.Compile(emailPattern)
	email := ""
	if err == nil {
		if emailRegex.MatchString(username) {
			email = username
		}
	}
	phonePattern := `^(\+\d{1,2}\s?)?(\(\d{3}\)|\d{3})[-.\s]?\d{3}[-.\s]?\d{4}$`
	phoneRegex := regexp.MustCompile(phonePattern)
	phone := ""
	if err == nil {
		if phoneRegex.MatchString(username) {
			phone = username
		}
	}
	cred := Credential{Username: username, Password: password, Email: email, Phone: phone}
	cred.CredStore = credStore
	return &cred
}

func (cred *Credential) SetCredential(username, password string) error {
	cred.Password = password
	cred.Username = username
	return nil
}

func (cred *Credential) GetString(field string) (string, error) {
	mapCred := mcli_utils.StructToMapStringValues(*cred)
	fieldValue, ok := mapCred[field]
	if ok {
		return *fieldValue, nil
	}
	return "", errors.New("no such field")
}

type Session struct {
	CookieName string
	Token      string
	Value      interface{}
	Expire     time.Duration
	Store      mcli_interface.KVStorer
}

func NewSession(cookieName string, s mcli_interface.KVStorer) *Session {
	session := Session{CookieName: cookieName, Store: s}
	return &session
}

func (session *Session) SetToken(sessionToken string) (string, error) {
	if sessionToken == "" {
		sessionToken = uuid.New().String()
	}
	if sessionToken == "" {
		return sessionToken, fmt.Errorf("empty session token generated")
	}
	session.Token = sessionToken
	return sessionToken, nil
}

func (session *Session) SetValue(v interface{}) error {
	session.Value = v
	return nil
}

func (session *Session) Authenticate(cred Credential) (bool, error) {
	ok, err := cred.CredStore.CheckPassword(cred.Username, cred.Password)
	// fmt.Println("Authenticate:", ok, err)
	if !ok {
		return ok, fmt.Errorf("authenticate error: %v", err)
	}
	err = session.Store.SetRecordEx(session.Token, cred.Username, int(session.Expire), "session-list")

	if err != nil {
		return false, fmt.Errorf("store session token error: %v", err)
	}
	return true, nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//  TODO: check routes and improove cookie name resolution
	clearAuthenticatedCookie(w, &Session{CookieName: cookieName})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
