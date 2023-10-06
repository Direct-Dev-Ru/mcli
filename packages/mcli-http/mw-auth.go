package mclihttp

import (
	"context"
	"net/http"
	"strings"
	"time"

	// "github.com/rs/zerolog/pkgerrors"
	mcli_interface "mcli/packages/mcli-interface"
)

type Auth struct {
	User        *Credential
	Inner       http.Handler
	isEncCookie bool
	userStore   mcli_interface.CredentialStorer
	kvStore     mcli_interface.KVStorer
}

type ContextKey string

func NewAuth(userStore mcli_interface.CredentialStorer, kvStore mcli_interface.KVStorer, isEnc bool) *Auth {

	return &Auth{userStore: userStore, kvStore: kvStore, isEncCookie: isEnc}
}

func (auth *Auth) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	_, ok := req.Context().Value(ContextKey("router")).(*Router)
	if !ok {
		http.Error(res, "StatusUnauthorized - no router in context", http.StatusUnauthorized)
		return
	}
	// fmt.Println("is router in ctx", ok, req.URL.Path)

	// fmt.Println("checking password")
	// fmt.Println(r.CredentialStore.CheckPassword("admin", "userOk"))

	cookie, err := auth.GetCookie(req, cookieName)
	var ctx context.Context = req.Context()
	if err != nil {
		ctx = context.WithValue(ctx, ContextKey("isAuth"), false)
		ctx = context.WithValue(ctx, ContextKey("authUser"), nil)

	} else {
		// fmt.Println("session-token", cookie)

		rawUserName, ttl, err := auth.kvStore.GetRecordEx(cookie, "session-list")
		username := strings.TrimPrefix(strings.TrimSuffix(string(rawUserName), `"`), `"`)

		// fmt.Println(username, ttl, err)
		if ttl <= 0 && err != nil {
			http.Error(res, "StatusUnauthorized - no session found in store", http.StatusUnauthorized)
			return
		}

		user, err, ok := auth.userStore.GetUser(username)
		// fmt.Println(user, err, ok)
		if ok && err != nil {
			http.Error(res, "StatusUnauthorized - no user found in store", http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, ContextKey("isAuth"), true)
		ctx = context.WithValue(ctx, ContextKey("authUser"), user)
	}
	auth.Inner.ServeHTTP(res, req.WithContext(ctx))
}

func (auth *Auth) SetInnerHandler(next http.Handler) {
	auth.Inner = next
}

func (auth *Auth) GetCookie(r *http.Request, cName string) (string, error) {
	cookie, err := r.Cookie(cName)
	// fmt.Println(r.URL.Path, r.Cookies())
	if err != nil {
		return "", err
	}
	if auth.isEncCookie {
		var value string

		err = s.Decode(cookieName, cookie.Value, &value)
		return value, err
	}
	return cookie.Value, err
}

func (auth *Auth) SetCookie(w http.ResponseWriter, session *Session) error {
	var cookieValueToStore = session.Token
	var err error
	if auth.isEncCookie {
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
