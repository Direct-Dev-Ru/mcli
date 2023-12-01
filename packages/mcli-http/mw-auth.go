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
		http.Error(res, "Status Unauthorized: No router in Context", http.StatusUnauthorized)
		return
	}
	// fmt.Println("is router in ctx", ok, req.URL.Path)

	// fmt.Println("checking password")
	// fmt.Println(r.CredentialStore.CheckPassword("admin", "userOk"))

	var ctx context.Context = req.Context()
	cookie, err := auth.GetCookie(req, cookieName)
	// if no cookieName = "session-token" then sets noAuth in context
	if len(cookie) == 0 || err != nil {
		ctx = context.WithValue(ctx, ContextKey("IsAuth"), false)
		ctx = context.WithValue(ctx, ContextKey("AuthUser"), nil)

	} else {

		// fmt.Printf("%s cookie in context of route %s: %s\n", cookieName, req.URL, cookie)

		// getting user from kvStore
		sessionPrefix := HttpConfig.Server.Auth.SessionsRedisPrefix
		if sessionPrefix == "" {
			sessionPrefix = "session-list"
		}
		rawUserName, ttl, err := auth.kvStore.GetRecordEx(cookie, sessionPrefix)
		username := strings.TrimPrefix(strings.TrimSuffix(string(rawUserName), `"`), `"`)

		// fmt.Println(username, ttl, err)

		if ttl <= 0 && err != nil {
			http.Error(res, "status unauthorized. no session found in store", http.StatusUnauthorized)
			return
		}
		// TODO: if rest time less than 80% - redirect to prolongate route

		// getting user raw data from kvstore
		userRaw, err, ok := auth.userStore.GetUser(username)
		if ok && err != nil {
			http.Error(res, "status unauthorized. no user found in store", http.StatusUnauthorized)
			return
		}
		// convert to Credential datatype
		user, ok := userRaw.(*Credential)
		if !ok {
			http.Error(res, "status unauthorized. user bad type in store", http.StatusUnauthorized)
			return
		}
		user.Password = ""

		// fmt.Println("user to store in Context", user, err, ok)

		// processing different cases with user state

		if !user.Confirmed {
			http.Redirect(res, req, HttpConfig.GetFullUrl(HttpConfig.Server.Auth.SignUpConfirmRoute),
				http.StatusTemporaryRedirect)
			return
		}

		if user.Blocked {
			// if user blocked - redirect to signup route
			http.Redirect(res, req, HttpConfig.GetFullUrl(HttpConfig.Server.Auth.SignUpRoute),
				http.StatusTemporaryRedirect)
			return
		}

		if user.Expired {
			// if password has expired - redirect to change password route
			http.Redirect(res, req, HttpConfig.GetFullUrl(HttpConfig.Server.Auth.SignInChangeRoute),
				http.StatusTemporaryRedirect)
			return
		}

		ctx = context.WithValue(ctx, ContextKey("IsAuth"), true)
		ctx = context.WithValue(ctx, ContextKey("AuthUser"), user)
	}
	auth.Inner.ServeHTTP(res, req.WithContext(ctx))
}

func (auth *Auth) SetInnerHandler(next http.Handler) {
	auth.Inner = next
}

func (auth *Auth) GetCookie(r *http.Request, cName string) (string, error) {
	cookie, err := r.Cookie(cName)
	// fmt.Println(r.URL.Path, r.Cookies())
	if err == http.ErrNoCookie {
		return "", nil // Return an empty string if the cookie is not found
	} else if err != nil {
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
