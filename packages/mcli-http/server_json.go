package mclihttp

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
)

func handleJsonRequest(w http.ResponseWriter, request *http.Request) {
	url, err := request.URL.Parse(request.RequestURI)
	if err != nil {
		RenderErrorJSON(w, err)
		return
	}
	log.Println(url)
	q := url.Query()

	if path := q.Get("path"); path != "" {

		files, err := os.ReadDir(path)
		if err != nil {
			RenderErrorJSON(w, err)
			return
		}

		dirfiles := make([]string, 0, 10)
		for _, file := range files {
			if !file.IsDir() && !strings.HasSuffix(strings.ToLower(file.Name()), ".exe") {
				dirfiles = append(dirfiles, file.Name())
			}
		}
		RenderJSON(w, dirfiles, true)
	} else {
		RenderErrorJSON(w, errors.New("no path specified"))
	}

}
