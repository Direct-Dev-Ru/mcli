package mclihttp

import (
	"net/http"
	"strings"
)

type RouteType int

const (
	Equal RouteType = iota + 1
	Prefix
	Regexp
)

type HandleFunc func(http.ResponseWriter, *http.Request)

// Route
type Route struct {
	url       string
	pattern   string
	routeType RouteType
	method    string
	Handler   HandleFunc
}

func NewRoute(pattern string, routeType RouteType) *Route {
	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	return &Route{"", pattern, routeType, "", http.NotFound}
}

func NewRouteWithMethod(pattern string, routeType RouteType, method string) *Route {
	method = strings.ToUpper(method)
	if !(method == http.MethodGet || method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete) {
		method = ""
	}

	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	return &Route{"", pattern, routeType, method, http.NotFound}
}

func NewRouteWithHandler(pattern string, routeType RouteType, f HandleFunc) *Route {
	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	if f == nil {
		f = http.NotFound
	}
	return &Route{"", pattern, routeType, "", f}
}

func NewRouteWithHandlerAndMethod(pattern string, routeType RouteType, f HandleFunc, method string) *Route {
	method = strings.ToUpper(method)
	if !(method == http.MethodGet || method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete) {
		method = ""
	}

	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	if f == nil {
		f = http.NotFound
	}

	return &Route{"", pattern, routeType, method, f}
}

func (r *Route) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	reqMethod := req.Method
	if r.Handler != nil {
		// fmt.Println(req.Method, req.URL, r.Handler)
		if len(r.method) == 0 {
			r.Handler(res, req)
			return
		}
		if len(r.method) > 0 && r.method == reqMethod {
			r.Handler(res, req)
			return
		}
	} else {
		http.Error(res, "404 Handler Not Found", 404)
		return
	}
}
func (r *Route) SetHandler(f HandleFunc) http.Handler {
	r.Handler = f
	return r
}
