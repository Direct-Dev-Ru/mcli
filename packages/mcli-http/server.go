package mclihttp

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
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
	staticHandler http.Handler
	routes        []*Route
	middleware    []http.Handler
}

func NewRouter(sPath string, sPrefix string) *Router {
	sPath = strings.TrimPrefix(sPath, "./")
	sPath = strings.TrimSuffix(sPath, "/")
	sPrefix = strings.TrimPrefix(sPrefix, "/")
	sPrefix = strings.TrimSuffix(sPrefix, "/")
	var fileServer http.Handler
	if !(len(sPath) == 0 || len(sPrefix) == 0) {
		fileServer = http.FileServer(http.Dir("./" + sPath))
	}

	return &Router{sPath: sPath, sPrefix: sPrefix, staticHandler: fileServer,
		middleware: make([]http.Handler, 0, 3), routes: make([]*Route, 0, 3)}
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

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
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

	// switch {
	// case strings.HasPrefix(req.URL.Path, "/"+r.sPrefix+"/"):
	// 	http.StripPrefix("/"+r.sPrefix, r.staticHandler).ServeHTTP(res, req)
	// case req.URL.Path == "/":
	// 	staticPath := "./" + r.sPath
	// 	if strings.HasPrefix(r.sPath, "/") {
	// 		// given absolute path to static content
	// 		staticPath = r.sPath
	// 	}
	// 	mainPagePath := ""
	// 	mainPagePathCandidate := staticPath + "/index.html"
	// 	if _, err := os.Stat(mainPagePathCandidate); err != nil {
	// 		mainPagePathCandidate = staticPath + "/html/index.html"
	// 		if _, err := os.Stat(mainPagePathCandidate); err != nil {
	// 			mainPagePathCandidate = ""
	// 		}
	// 	}
	// 	mainPagePath = mainPagePathCandidate

	// 	if len(mainPagePath) > 0 {
	// 		http.ServeFile(res, req, mainPagePath)
	// 	} else {
	// 		http.Error(res, "404 Not Found Root Index.html", 404)
	// 	}
	// case req.URL.Path == "/service/exit":
	// 	r.SetRouteHandler(req.URL.Path, "/service/exit", func(res http.ResponseWriter, req *http.Request) {
	// 		fmt.Println("Server is shutting down")
	// 		fmt.Fprint(res, "Server is shutting down")
	// 		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	// 	}).ServeHTTP(res, req)
	// case req.URL.Path == "/echo":
	// 	r.SetRouteHandler(req.URL.Path, "/echo", http_echo).ServeHTTP(res, req)
	// case strings.HasPrefix(req.URL.Path, "/json/listfiles"):
	// 	r.SetRouteHandler(req.URL.Path, "/json/listfiles", handleJsonRequest).ServeHTTP(res, req)
	// default:
	// 	http.Error(res, "404 Not Found", 404)
	// }
}

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
