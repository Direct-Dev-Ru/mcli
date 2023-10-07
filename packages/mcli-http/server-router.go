package mclihttp

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	mcli_interface "mcli/packages/mcli-interface"

	"github.com/rs/zerolog"
)

// Router
type Router struct {
	sPath           string
	sPrefix         string
	sBaseURL        string
	infoLog         zerolog.Logger
	errorLog        zerolog.Logger
	staticHandler   http.Handler
	routes          []*Route
	middleware      []mcli_interface.Middleware
	finalHandler    http.Handler
	KVStore         mcli_interface.KVStorer
	CredentialStore mcli_interface.CredentialStorer
	Cache           map[string]mcli_interface.Cacher
}

type RouterOptions struct {
	BaseUrl string
}

func NewRouter(sPath string, sPrefix string, iLog zerolog.Logger, Elogger zerolog.Logger, opts RouterOptions) *Router {
	sPath = strings.TrimPrefix(sPath, "./")
	sPath = strings.TrimSuffix(sPath, "/")
	sPrefix = strings.TrimPrefix(sPrefix, "/")
	sPrefix = strings.TrimSuffix(sPrefix, "/")

	baseURL := opts.BaseUrl
	baseURL = strings.TrimPrefix(baseURL, "/")

	var fileServer http.Handler

	if !(len(sPath) == 0 || len(sPrefix) == 0) {
		fileServerResultPath := sPath
		if !strings.HasPrefix(sPath, "/") {
			fileServerResultPath = "./" + sPath
		}
		// fmt.Println(fileServerResultPath)
		fileServer = http.FileServer(http.Dir(fileServerResultPath))
	}

	return &Router{infoLog: iLog, errorLog: Elogger, sPath: sPath, sPrefix: sPrefix, staticHandler: fileServer, sBaseURL: baseURL,
		middleware: make([]mcli_interface.Middleware, 0, 3), routes: make([]*Route, 0, 3)}
}

func (r *Router) PrintRoutes() {

	for _, route := range r.routes {
		fmt.Println(route)
	}
}

func (r *Router) Use(mw mcli_interface.Middleware) error {
	r.middleware = append(r.middleware, mw)
	return nil
}

func (r *Router) ConstructFinalHandler() error {

	r.finalHandler = http.HandlerFunc(r.innerHandler)

	if len(r.middleware) > 0 {
		var currentMw mcli_interface.Middleware
		for i := len(r.middleware) - 1; i >= 0; i-- {
			currentMw = r.middleware[i]
			currentMw.SetInnerHandler(r.finalHandler)
			r.finalHandler = currentMw
		}
	}
	return nil
}

func (r *Router) injectToContext(next http.HandlerFunc, keyCtx string, valueCtx interface{}) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Add a value to the context
		// userID := "user123"
		var ctx = req.Context()
		if valueCtx != nil {
			ctx = context.WithValue(req.Context(), ContextKey(keyCtx), valueCtx)
		}
		// Call the next handler with the updated context
		next(res, req.WithContext(ctx))
	}
}

func (r *Router) AddRoute(route *Route) error {
	if route.Handler == nil {
		return fmt.Errorf("route handler is nil")
	}
	if route.pattern == "/" && len(r.sBaseURL) > 0 {
		// add duplicate root route
		rootRouteClone := NewRoute("/"+r.sBaseURL, Equal)
		rootRouteClone.Handler = route.Handler
		r.routes = append(r.routes, rootRouteClone)
	}
	route.pattern = r.getResultPattern(route.pattern)
	r.routes = append(r.routes, route)
	return nil
}

func (r *Router) getResultPattern(partial string) string {
	if len(r.sBaseURL) == 0 {
		return partial
	}
	partial = strings.TrimPrefix(partial, "/")
	return "/" + r.sBaseURL + "/" + partial
}

func (r *Router) AddRouteWithHandler(pattern string, routeType RouteType, f HandleFunc) error {
	if f == nil {
		f = http.NotFound
	}
	route := NewRouteWithHandler(r.getResultPattern(pattern), routeType, f)
	r.routes = append(r.routes, route)
	return nil
}

func (r *Router) innerHandler(res http.ResponseWriter, req *http.Request) {
	reqPath := strings.TrimSpace(req.URL.Path)

	// fmt.Println(reqPath)
	// serving static assets
	if strings.HasPrefix(reqPath, "/"+r.sPrefix+"/") && r.staticHandler != nil {
		http.StripPrefix("/"+r.sPrefix, r.staticHandler).ServeHTTP(res, req)
		return
	}

	if reqPath == "/favicon.ico" {
		http.Redirect(res, req, r.sPrefix+"/favicon.ico", http.StatusFound)
		return
	}

	// serving routes in router
	for _, route := range r.routes {

		switch route.routeType {
		case Equal:
			if reqPath == route.pattern {
				route.ServeHTTP(res, req)
				return
			}
		case Prefix:

			if strings.HasPrefix(reqPath, route.pattern) {
				// fmt.Println(reqPath, route.pattern)
				// fmt.Println(route.Handler)
				route.ServeHTTP(res, req)
				return
			}
		case Regexp:
			re, err := regexp.Compile(route.pattern)
			if err == nil {
				if re.MatchString(reqPath) {
					route.ServeHTTP(res, req)
					return
				}
			}
		default:
			http.Error(res, "404 Not Found", 404)
		}
	}
	http.Error(res, "404 Not Found", 404)

}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if r.finalHandler == nil {
		r.ConstructFinalHandler()
		// http.HandlerFunc(r.innerHandler).ServeHTTP(res, req)
	}
	r.injectToContext(r.finalHandler.ServeHTTP, "router", r).ServeHTTP(res, req)
	// r.finalHandler.ServeHTTP(res, req)
}
