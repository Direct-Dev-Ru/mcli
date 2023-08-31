package mclihttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	sc "github.com/gorilla/securecookie"
)

//	var users = map[string]string{
//		"user1": "pwd1",
//		"user2": "pwd2",
//	}
var cookieName string = "session-token"
var (
	timeExeed     = 36000
	s             *sc.SecureCookie
	encodeCookies = false
)

func SetSecretCookieOptions(doEncoding bool, cookieHash, cookieBlock []byte) {
	encodeCookies = doEncoding
	s = sc.New(cookieHash, cookieBlock)
}

type Credential struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	FirstName   string   `json:"first-name"`
	LastName    string   `json:"last-name"`
	Email       string   `json:"email"`
	BackupEmail string   `json:"backup-email"`
	Phone       string   `json:"phone"`
	BackupPhone string   `json:"backup-phone"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
	CredStore   CredentialStorer
}

func NewCredential(username, password string, credStore CredentialStorer) *Credential {
	cred := Credential{Username: username, Password: password}
	cred.CredStore = credStore
	return &cred
}

func (cred *Credential) SetCredential(username, password string) error {
	cred.Password = password
	cred.Username = username
	return nil
}

type Session struct {
	CookieName string
	Token      string
	Value      interface{}
	Expire     time.Duration
	Store      KVStorer
}

func NewSession(cookieName string, s KVStorer) *Session {
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
	clearAuthenticatedCookie(w, &Session{CookieName: "session-token"})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
