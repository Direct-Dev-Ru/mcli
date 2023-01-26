package mclihttp

import (
	"errors"
	"net/http"
)

func HandleJsonRequest(res http.ResponseWriter, req *http.Request) {
	url, err := req.URL.Parse(req.RequestURI)
	if err != nil {
		RenderErrorJSON(res, err)
		return
	}
	q := url.Query()

	if data := q.Get("data"); data != "" {
		RenderJSON(res, data, true)
	} else {
		RenderErrorJSON(res, errors.New("no data specified"))
	}
}
