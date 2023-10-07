package mclihttp

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Http_Echo(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(writer, "Method: %v\n", request.Method)
	for header, vals := range request.Header {
		fmt.Fprintf(writer, "Header: %v: %v\n", header, vals)
	}
	fmt.Fprintln(writer, "----------Context-------------")
	isAuth, ok := request.Context().Value(ContextKey("IsAuth")).(bool)
	if ok {
		fmt.Fprintf(writer, "Is Authenticated: %v\n", isAuth)
	}

	user, ok := request.Context().Value(ContextKey("AuthUser")).(*Credential)
	if ok {
		fmt.Fprintf(writer, "Authenticated user.Username: %v\n", user.Username)
		fmt.Fprintf(writer, "Authenticated user.Roles: %v\n", user.Roles)
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
}
