package mclihttp

import (
	"encoding/json"
	"net/http"
)

func RenderJSON(w http.ResponseWriter, v interface{}, wrap bool) {
	type SuccessResponse struct {
		IsError bool        `json:"iserror"`
		Error   string      `json:"error"`
		Payload interface{} `json:"payload"`
	}
	var send []byte
	var err error
	if wrap {
		send, err = json.Marshal(SuccessResponse{false, "", v})
	} else {
		send, err = json.Marshal(v)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(send)

}

func RenderErrorJSON(w http.ResponseWriter, e error) {
	type ErrorResponse struct {
		IsError bool   `json:"iserror"`
		Error   string `json:"error"`
	}
	errRes := ErrorResponse{IsError: true, Error: e.Error()}
	json, err := json.Marshal(errRes)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
