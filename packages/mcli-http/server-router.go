package mclihttp

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	mcli_interface "mcli/packages/mcli-interface"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/rs/zerolog"
)

// Router
type Router struct {
	sPath         string
	sPrefix       string
	sBaseURL      string
	infoLog       zerolog.Logger
	errorLog      zerolog.Logger
	staticHandler http.Handler
	routes        []*Route
	// RouteType -- Method -- Pattern
	mapRoutes       map[RouteType]map[string]map[string]*Route
	middleware      []mcli_interface.Middleware
	finalHandler    http.Handler
	KVStore         mcli_interface.KVStorer
	CredentialStore mcli_interface.CredentialStorer
	Cache           mcli_interface.Cacher
}

type RouterOptions struct {
	BaseUrl         string
	KVStore         mcli_interface.KVStorer
	CredentialStore mcli_interface.CredentialStorer
}

func NewRouter(sPath string, sPrefix string, iLog zerolog.Logger, Elogger zerolog.Logger, opts *RouterOptions) *Router {
	sPath = strings.TrimPrefix(sPath, "./")
	sPath = strings.TrimSuffix(sPath, "/")
	sPrefix = strings.TrimPrefix(sPrefix, "/")
	sPrefix = strings.TrimSuffix(sPrefix, "/")

	var fileServer http.Handler

	if !(len(sPath) == 0 || len(sPrefix) == 0) {
		fileServerResultPath := sPath
		if !strings.HasPrefix(sPath, "/") {
			fileServerResultPath = "./" + sPath
		}
		// fmt.Println(fileServerResultPath)
		fileServer = http.FileServer(http.Dir(fileServerResultPath))
	}
	mapRoutes := make(map[RouteType]map[string]map[string]*Route, 3)
	mapRoutes[Equal] = make(map[string]map[string]*Route, 0)
	mapRoutes[Prefix] = make(map[string]map[string]*Route, 0)
	mapRoutes[Regexp] = make(map[string]map[string]*Route, 0)
	baseURL := ""
	router := Router{infoLog: iLog, errorLog: Elogger, sPath: sPath, sPrefix: sPrefix, staticHandler: fileServer,
		sBaseURL: baseURL, middleware: make([]mcli_interface.Middleware, 0, 3), routes: make([]*Route, 0, 3),
		mapRoutes: mapRoutes}
	if opts != nil {
		baseURL = strings.TrimSpace(opts.BaseUrl)
		baseURL = strings.TrimPrefix(baseURL, "/")
		if len(opts.BaseUrl) > 0 {
			router.sBaseURL = baseURL
		}
		if opts.KVStore != nil {
			router.KVStore = opts.KVStore
		}
		if opts.CredentialStore != nil {
			router.CredentialStore = opts.CredentialStore
		}
	}
	router.Cache = mcli_utils.NewCCache(600, 100, func(params ...interface{}) (interface{}, error) {
		if len(params) == 0 {
			return nil, fmt.Errorf("no params provided")
		}
		if len(params) == 1 {
			valToProcess, ok := params[0].(*Route)
			if ok {
				return valToProcess, nil
			}
			return nil, fmt.Errorf("wrong type parameter for cache function")
		}
		return nil, fmt.Errorf("too many parameters for cache function")
	})

	return &router
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
			ctx = context.WithValue(req.Context(), mcli_interface.ContextKey(keyCtx), valueCtx)
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
	// fmt.Println(route.pattern)
	r.routes = append(r.routes, route)

	switch route.routeType {
	case Equal:
		// if pattern contains :params parts in path, f.e. /baseurl/path1/:section/:post
		if strings.Contains(route.pattern, ":") {
			re, err := regexp.Compile(`:([^/]+)`)
			if err != nil {
				return err
			}
			regexPattern := re.ReplaceAllStringFunc(route.pattern, func(match string) string {
				// Convert :param to (?P<param>[^/]+)
				paramName := match[1:]
				return fmt.Sprintf(`(?P<%s>[^/]+)`, paramName)
			})

			regExpPattern, err := regexp.Compile("^" + regexPattern + "$")
			if err != nil {
				return err
			}
			route.regexp = regExpPattern
			route.pattern = "@UseRegExp->" + route.pattern
		}
		r.setRouterMapsByMethod(route)
	case Prefix:
		r.setRouterMapsByMethod(route)
	case Regexp:
		rExp, err := regexp.Compile(route.pattern)
		if err != nil {
			return err
		}
		route.regexp = rExp
		r.setRouterMapsByMethod(route)
	}

	return nil
}

func (r *Router) setRouterMapsByMethod(route *Route) {
	switch route.method {
	case http.MethodGet:
		r.mapRoutes[route.routeType][http.MethodGet][route.pattern] = route
	case http.MethodPost:
		r.mapRoutes[route.routeType][http.MethodPost][route.pattern] = route
	case http.MethodPut:
		r.mapRoutes[route.routeType][http.MethodPut][route.pattern] = route
	case http.MethodDelete:
		r.mapRoutes[route.routeType][http.MethodDelete][route.pattern] = route
	default:
		if _, ok := r.mapRoutes[route.routeType]["General"]; !ok {
			r.mapRoutes[route.routeType]["General"] = make(map[string]*Route)
		}
		r.mapRoutes[route.routeType]["General"][route.pattern] = route
	}
}

func (r *Router) getResultPattern(partial string) string {
	if len(r.sBaseURL) == 0 {
		return partial
	}
	return HttpConfig.GetFullUrl(partial, r.sBaseURL)
}

func (r *Router) AddRouteWithHandler(pattern string, routeType RouteType, f HandlerFunc) error {
	if f == nil {
		f = http.NotFound
	}
	route := NewRouteWithHandler(r.getResultPattern(pattern), routeType, f)
	// r.routes = append(r.routes, route)
	err := r.AddRoute(route)
	if err != nil {
		return err
	}
	return nil
}

// this handler fires after all middlewares
func (r *Router) innerHandler(res http.ResponseWriter, req *http.Request) {
	reqPath := strings.TrimSpace(req.URL.Path)
	if isHas := strings.HasSuffix(reqPath, "/"); isHas {
		reqPath = strings.TrimSuffix(reqPath, "/")
	}
	reqPathSlash := strings.TrimSpace(reqPath) + "/"
	reqPaths := []string{reqPath, reqPathSlash}

	reqMethod := req.Method
	// fmt.Println("innerHandler:", reqPaths, reqMethod)

	// serving static assets
	if strings.HasPrefix(reqPath, "/"+r.sPrefix+"/") && r.staticHandler != nil {
		http.StripPrefix("/"+r.sPrefix, r.staticHandler).ServeHTTP(res, req)
		return
	}

	if reqPath == "/favicon.ico" {
		http.Redirect(res, req, r.sPrefix+"/favicon.ico", http.StatusFound)
		return
	}

	if HttpConfig.Server.RouterV2 {
		// V2 routing
		// Equal Routes
		for _, rPath := range reqPaths {
			// first we need to try routes cache
			r.infoLog.Trace().Msgf("resolve route for path %s", rPath)
			iRoute, err := r.Cache.Get(rPath)

			if err == nil && iRoute != nil {
				if route, ok := iRoute.(*Route); ok {
					r.infoLog.Trace().Msgf("for path %s get route %v from cache ", rPath, []*Route{route})
					route.ServeHTTP(res, req)
					return
				}
			}

			// process equal type of routes
			generalEqualRoutes := r.mapRoutes[Equal]["General"]
			methodEqualRoutes := r.mapRoutes[Equal][reqMethod]
			mapEqualArray := []map[string]*Route{methodEqualRoutes, generalEqualRoutes}

			for i := 0; i < len(mapEqualArray); i++ {
				mapRoutes := mapEqualArray[i]
				if len(mapRoutes) > 0 {
					if route, exists := mapRoutes[rPath]; exists {
						r.infoLog.Trace().Msgf("path %s try to set cache for equal route %v", rPath, []*Route{route})

						_, err := r.Cache.Set(reqPath, route)
						if err != nil {
							r.errorLog.Err(err).Msgf("set route cache for path %s has fault:", rPath)
						}
						route.ServeHTTP(res, req)
						return
					}
					// search for regexp equal routes
					for _, route := range mapRoutes {
						// continue if there are no specific prefix
						if !strings.HasPrefix(route.pattern, "@UseRegExp->") {
							continue
						}
						//  if url path matches regexp calculated during route adding
						if match, reqParams := route.matchRoute(rPath); match {
							ctx := context.WithValue(req.Context(), mcli_interface.ContextKey("reqParams"), reqParams)
							// we do not use cache in this case cause we must parse request parameters
							// and set them to req.ctx
							// TODO: make caching route with parameters - wrap in some struct
							// in handler we access them for example such way:
							// article := req.Context().Value(ctxKeys("reqParams")).(map[string]string)["section"]
							// post := req.Context().Value(ctxKeys("reqParams")).(map[string]string)["post"]

							route.ServeHTTP(res, req.WithContext(ctx))
							return
						}
					}
				}
			}
		}
		// Prefix Routes
		for _, rPath := range reqPaths {
			generalPrefixRoutes := r.mapRoutes[Prefix]["General"]
			methodPrefixRoutes := r.mapRoutes[Prefix][reqMethod]
			mapPrefixArray := []*map[string]*Route{&methodPrefixRoutes, &generalPrefixRoutes}

			for i := 0; i < len(mapPrefixArray); i++ {
				mapPrefixRoutes := mapPrefixArray[i]

				for _, prefixRoute := range *mapPrefixRoutes {
					if strings.HasPrefix(rPath, prefixRoute.pattern) {
						// fmt.Println("try to set cache for ", []*Route{prefixRoute})
						_, err := r.Cache.Set(reqPath, prefixRoute)
						if err != nil {
							r.errorLog.Err(err).Msgf("set route cache for path %s has fault:", rPath)
						}
						prefixRoute.ServeHTTP(res, req)
						return
					}
				}
			}
		}
		// RegExp Routes
		for _, rPath := range reqPaths {
			generalRegExpRoutes := r.mapRoutes[Regexp]["General"]
			methodRegExpRoutes := r.mapRoutes[Regexp][reqMethod]
			mapRegExpArray := []map[string]*Route{methodRegExpRoutes, generalRegExpRoutes}

			for i := 0; i < len(mapRegExpArray); i++ {
				mapRegExpRoutes := mapRegExpArray[i]
				for _, regExpRoute := range mapRegExpRoutes {
					// fmt.Println(regExpRoute.pattern, regExpRoute.regexp)
					if regExpRoute.regexp != nil {
						if match, reqParamsArray := regExpRoute.matchRouteParamArray(rPath); match {
							ctx := context.WithValue(req.Context(), mcli_interface.ContextKey("reqParamArray"), reqParamsArray)
							// we do not use cache in this case cause we must parse request parameters
							// and set them to req.ctx
							// TODO: make caching route with parameters - wrap in some struct
							// in handler we access them for example such way:
							// article := req.Context().Value(mcli_interface.ContextKey("reqParamsArray")).([]string)[0]
							// post := req.Context().Value(mcli_interface.ContextKey("reqParamsArray")).([]string)[1]

							regExpRoute.ServeHTTP(res, req.WithContext(ctx))
							return
						}
					}
				}
			}
		}
		http.Error(res, "404 Not Found", 404)
	} else {
		// serving routes in router
		for _, route := range r.routes {
			// fmt.Println(route)
			// fmt.Println("innerHandler:", reqPaths, reqMethod)
			switch route.routeType {
			case Equal:
				for _, rPath := range reqPaths {
					if rPath == route.pattern {
						route.ServeHTTP(res, req)
						return
					}
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
	}

}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if r.finalHandler == nil {
		r.ConstructFinalHandler()
		// http.HandlerFunc(r.innerHandler).ServeHTTP(res, req)
	}
	r.injectToContext(r.finalHandler.ServeHTTP, "router", r).ServeHTTP(res, req)
	// r.finalHandler.ServeHTTP(res, req)
}
