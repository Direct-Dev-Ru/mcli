package mclihttp

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func init() {

	http.HandleFunc("/html",
		func(writer http.ResponseWriter, request *http.Request) {
			// ex, _ := os.Executable()

			// fmt.Println(runtime.Caller(0))
			http.ServeFile(writer, request, "./public/index.html")
		})

	http.HandleFunc("/service/exit",
		func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintf(writer, "Server is shutting down %s... /n", nil)
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
