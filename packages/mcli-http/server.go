package mclihttp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func InitMainRoutes(sPath string, sPrefix string) {

	sPath = strings.TrimPrefix(sPath, "./")
	sPath = strings.TrimSuffix(sPath, "/")
	sPrefix = strings.TrimPrefix(sPrefix, "/")
	sPrefix = strings.TrimSuffix(sPrefix, "/")

	fileServer := http.FileServer(http.Dir("./" + sPath))
	http.Handle("/"+sPrefix+"/", http.StripPrefix("/"+sPrefix, fileServer))

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
