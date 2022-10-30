package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	//"encoding/json"
	//"strings"
	//"net/url"
	//"net/http/cookiejar"
	//"fmt"
)

func readBody(rc io.ReadCloser) (string, error) {
	data, err := io.ReadAll(rc)
	if err == nil {
		defer rc.Close()
		strBody, err := url.QueryUnescape(string(data))
		if err == nil {
			return strBody, nil
		}
	}
	return "", err
}

func main() {
	go http.ListenAndServe(":5000", nil)
	time.Sleep(time.Second * 1)

	var response *http.Response
	var err error

	baseURI := "http://localhost:5000"
	baseURI = "localhost:33333"
	// response, err = http.Get("http://localhost:5000/html")
	// if err == nil && response.StatusCode == http.StatusOK {

	// 	data, err := io.ReadAll(response.Body)
	// 	if err == nil {
	// 		defer response.Body.Close()
	// 		strBody := string(data)
	// 		// os.Stdout.Write(data)
	// 		fmt.Println(strBody)
	// 	}

	// } else {
	// 	Printfln("Error: %v, Status Code: %v", err.Error(),
	// 		response.StatusCode)
	// }

	// response, err = http.Get("http://localhost:5000/json/listfiles?path=c:\\temp")
	// if err == nil && response.StatusCode == http.StatusOK {
	// 	defer response.Body.Close()
	// 	var data interface{}
	// 	err = json.NewDecoder(response.Body).Decode(&data)
	// 	if err == nil {
	// 		Printfln("Decoded value: %v", data)
	// 	} else {
	// 		Printfln("Decode error: %v", err.Error())
	// 	}
	// } else {
	// 	Printfln("Error: %v, Status Code: %v", err.Error(),
	// 		response.StatusCode)
	// }

	// formData := map[string][]string{
	// 	"name":     {"Футболка"},
	// 	"category": {"Спортивная одежда"},
	// 	"price":    {"569"},
	// }

	// response, err = http.PostForm(fmt.Sprintf("%s/echo", baseURI), formData)
	// if err == nil && response.StatusCode == http.StatusOK {
	// 	strBody, _ := readBody(response.Body)
	// 	fmt.Println(strBody)
	// } else {
	// 	Printfln("Error: %v, Status Code: %v", err.Error(),
	// 		response.StatusCode)
	// }

	payload := struct {
		Login    string
		Password string
	}{
		Login:    "kuznetcovay",
		Password: "dsasdsafdsadfwe",
	}

	var builder strings.Builder
	err = json.NewEncoder(&builder).Encode(payload)

	// if err == nil {
	// 	response, err = http.Post(fmt.Sprintf("%s/echo", baseURI),
	// 		"application/json", strings.NewReader(builder.String()))
	// }
	// if err == nil && response.StatusCode == http.StatusOK {
	// 	strBody, _ := readBody(response.Body)
	// 	fmt.Println(strBody)
	// } else {
	// 	Printfln("Error: %v, Status Code: %v", err.Error(),
	// 		response.StatusCode)
	// }

	// if err == nil {
	// 	req, err := http.NewRequest(http.MethodPost,
	// 		fmt.Sprintf("%s/echo", baseURI),
	// 		io.NopCloser(strings.NewReader(builder.String())))
	// 	if err == nil {
	// 		req.Header["Content-Type"] = []string{"application/json"}
	// 		response, err = http.DefaultClient.Do(req)

	// 		if err == nil && response.StatusCode == http.StatusOK {

	// 			strBody, _ := readBody(response.Body)
	// 			fmt.Println(strBody)
	// 			// fmt.Println("----")
	// 			// fmt.Println(response.Header)
	// 		} else {
	// 			Printfln("Error: %v, Status Code: %v", err.Error(),
	// 				response.StatusCode)
	// 		}
	// 	}
	// }
	if err == nil {
		req, err := http.NewRequest(http.MethodPost,
			baseURI,
			io.NopCloser(strings.NewReader(builder.String())))
		if err == nil {

			response, err = http.DefaultClient.Do(req)
			strBody, _ := readBody(response.Body)
			fmt.Println(strBody, err)

		}
	}
	time.Sleep(time.Second * 1000)
}
