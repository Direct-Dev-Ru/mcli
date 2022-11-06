package mclihttp

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func handleJsonRequest(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	url, err := request.URL.Parse(request.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(url)
	q := url.Query()

	if path := q.Get("path"); path != "" {
		// expath, _ := os.Executable()
		log.Println(path)
		files, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}
		dirfiles := make([]string, 0, 10)

		for _, file := range files {
			if !file.IsDir() && !strings.HasSuffix(strings.ToLower(file.Name()), ".exe") {
				dirfiles = append(dirfiles, file.Name())
			}
		}

		json.NewEncoder(w).Encode(dirfiles)
	} else {
		json.NewEncoder(w).Encode("no path specified")
	}

}
