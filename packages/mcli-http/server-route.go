package mclihttp

import (
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

type HandlerFunc func(http.ResponseWriter, *http.Request)

// Route
type Route struct {
	url       string
	pattern   string
	regexp    *regexp.Regexp
	routeType RouteType
	method    string
	Handler   HandlerFunc
}

func NewRoute(pattern string, routeType RouteType) *Route {
	if routeType <= 0 || routeType > 2 {
		routeType = Equal
	}
	return &Route{"", pattern, nil, routeType, "", http.NotFound}
}

func NewRouteWithMethod(pattern string, routeType RouteType, method string) *Route {
	method = strings.ToUpper(method)
	if !(method == http.MethodGet || method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete) {
		method = ""
	}

	route := NewRoute(pattern, routeType)
	route.method = method
	return route
}

func NewRouteWithHandler(pattern string, routeType RouteType, f HandlerFunc) *Route {
	if routeType < 1 || routeType > 3 {
		routeType = Equal
	}
	if f == nil {
		f = http.NotFound
	}
	return &Route{"", pattern, nil, routeType, "", f}
}

func NewRouteWithMethodAndHandler(pattern string, routeType RouteType, method string, f HandlerFunc) *Route {
	method = strings.ToUpper(method)
	if !(method == http.MethodGet || method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete) {
		method = ""
	}

	route := NewRouteWithHandler(pattern, routeType, f)
	route.method = method
	return route
}

func (r *Route) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	reqMethod := req.Method
	if r.Handler != nil {
		// fmt.Println(req.Method, req.URL, r.Handler)
		if len(r.method) > 0 && r.method == reqMethod {
			r.Handler(res, req)
			return
		}
		if len(r.method) == 0 {
			r.Handler(res, req)
			return
		}
	} else {
		http.Error(res, "404 Handler Not Found", 404)
		return
	}
}

func (r *Route) SetHandler(f HandlerFunc) http.Handler {
	r.Handler = f
	return r
}

func (r *Route) SetMethod(method string) http.Handler {
	method = strings.ToUpper(method)
	if !(method == http.MethodGet || method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete) {
		method = ""
	}
	r.method = method
	return r
}

func (r *Route) matchRoute(path string) (bool, map[string]string) {
	if r.regexp == nil {
		return false, nil
	}
	pattern := r.regexp

	matches := pattern.FindStringSubmatch(path)
	if matches == nil {
		return false, nil
	}

	params := make(map[string]string)
	for i, name := range pattern.SubexpNames()[1:] {
		//fmt.Println(matches[i+1])
		params[name] = matches[i+1]
	}

	return true, params
}

func (r *Route) matchRouteParamArray(path string) (bool, []string) {
	if r.regexp == nil {
		return false, nil
	}
	pattern := r.regexp

	matches := pattern.FindStringSubmatch(path)
	if matches == nil {
		return false, nil
	}

	params := make([]string, 0)
	for i, match := range matches {
		if i == 0 {
			continue
		}
		// fmt.Println(match)
		params = append(params, match)
	}

	return true, params
}
