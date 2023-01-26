package mclihttp

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

type RouteType int

const (
	Equal RouteType = iota + 1
	Prefix
	Regexp
)

type HandleFunc func(http.ResponseWriter, *http.Request)
type Route struct {
	url       string
	pattern   string
	routeType RouteType
	Handler   HandleFunc
}

func NewRoute(pattern string, routeType RouteType) *Route {
	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	return &Route{"", pattern, routeType, http.NotFound}
}

func NewRouteWithHandler(pattern string, routeType RouteType, f HandleFunc) *Route {
	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	if f == nil {
		f = http.NotFound
	}
	return &Route{"", pattern, routeType, f}
}

func (r *Route) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if r.Handler != nil {
		r.Handler(res, req)
	} else {
		http.Error(res, "404 Handler Not Found", 404)
	}
}
func (r *Route) SetHandler(f HandleFunc) http.Handler {
	r.Handler = f
	return r
}

type Router struct {
	sPath         string
	sPrefix       string
	infoLog       zerolog.Logger
	errorLog      zerolog.Logger
	staticHandler http.Handler
	routes        []*Route
	middleware    []Middleware
	finalHandler  http.Handler
}

func NewRouter(sPath string, sPrefix string, iLog zerolog.Logger, eLog zerolog.Logger) *Router {
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

	return &Router{infoLog: iLog, errorLog: eLog, sPath: sPath, sPrefix: sPrefix, staticHandler: fileServer,
		middleware: make([]Middleware, 0, 3), routes: make([]*Route, 0, 3)}
}

func (r *Router) Use(mw Middleware) error {
	r.middleware = append(r.middleware, mw)
	return nil
}

func (r *Router) ConstructFinalHandler() error {

	r.finalHandler = http.HandlerFunc(r.innerHandler)

	if len(r.middleware) > 0 {
		var currentMw Middleware
		for i := len(r.middleware) - 1; i >= 0; i-- {
			currentMw = r.middleware[i]
			currentMw.SetInnerHandler(r.finalHandler)
			r.finalHandler = currentMw
		}
	}
	return nil
}

func (r *Router) AddRoute(route *Route) error {
	if route.Handler == nil {
		return fmt.Errorf("route handler is nil")
	}
	r.routes = append(r.routes, route)
	return nil
}

func (r *Router) AddRouteWithHandler(pattern string, routeType RouteType, f HandleFunc) error {
	if f == nil {
		f = http.NotFound
	}
	route := NewRouteWithHandler(pattern, routeType, f)
	r.routes = append(r.routes, route)
	return nil
}

func (r *Router) innerHandler(res http.ResponseWriter, req *http.Request) {
	reqPath := strings.TrimSpace(req.URL.Path)
	// static paths
	if strings.HasPrefix(reqPath, "/"+r.sPrefix+"/") && r.staticHandler != nil {

		http.StripPrefix("/"+r.sPrefix, r.staticHandler).ServeHTTP(res, req)
		return
	}
	for _, route := range r.routes {
		switch route.routeType {
		case Equal:
			if reqPath == route.pattern {
				route.ServeHTTP(res, req)
				return
			}
		case Prefix:
			if strings.HasPrefix(reqPath, route.pattern) {
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
	if r.finalHandler != nil {
		r.finalHandler.ServeHTTP(res, req)
	} else {
		http.HandlerFunc(r.innerHandler).ServeHTTP(res, req)
	}
}

// if len(r.middleware) > 0 {
// 	var serve http.Handler = innerHandler
// 	for i := 0; i < len(r.middleware); i++ {
// 		mw := r.middleware[i]
// 		mwInner, ok := mw.(Middleware)
// 		if !ok {
// 			r.errorLog.Fatal().Msg("wrong middleware definition")
// 		}

// 		if i == 0 {
// 			mwInner.InnerHandler(http.HandlerFunc(r.innerHandler), mw)
// 		} else {
// 			mwInner.InnerHandler(serve, mw)
// 		}
// 		serve, ok = mwInner.(http.Handler)
// 		if !ok {
// 			r.errorLog.Fatal().Msg("wrong middleware definition")
// 		}
// 	}
// 	serve.ServeHTTP(res, req)
// }

/***
func InitMainRoutes(sPath string, sPrefix string) {

	sPath = strings.TrimPrefix(sPath, "./")
	sPath = strings.TrimSuffix(sPath, "/")
	sPrefix = strings.TrimPrefix(sPrefix, "/")
	sPrefix = strings.TrimSuffix(sPrefix, "/")

	// static Route
	fileServer := http.FileServer(http.Dir("./" + sPath))
	http.Handle("/"+sPrefix+"/", http.StripPrefix("/"+sPrefix, fileServer))
	// main Route
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// ex, _ := os.Executable()
		// fmt.Println(runtime.Caller(0))
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		http.ServeFile(writer, request, "./"+sPath+"/html/index.html")
	})

	http.HandleFunc("/service/exit",
		func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println("Server is shutting down")
			os.Exit(0)
		})

	http.HandleFunc("/json/listfiles", handleJsonRequest)

	http.HandleFunc("/echo",
		func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(writer, "Method: %v\n", request.Method)
			for header, vals := range request.Header {
				fmt.Fprintf(writer, "Header: %v: %v\n", header, vals)
			}
			fmt.Fprintln(writer, "-----------------------")

			defer request.Body.Close()
			data, err := io.ReadAll(request.Body)

			if err == nil {
				if len(data) == 0 {
					fmt.Fprintln(writer, "No body")
				} else {
					writer.Write(data)
				}
			} else {
				fmt.Fprintf(os.Stdout, "Error reading body: %v\n", err.Error())
			}
		})
}
*/
