package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// IndexHandler .
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world")
}

// ResponseData .
type ResponseData struct {
	Code    uint32          `json:"code"`
	Results json.RawMessage `json:"results"`
}

// JobResult .
type JobResult struct {
	ID     uint32 `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// GetAllJobsResultsHandler .
func GetAllJobsResultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		addCorsHeadersForOption(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	resultBytes, err := json.Marshal(getMockJobsResults())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	respBytes, err := json.Marshal(ResponseData{
		Code:    0,
		Results: resultBytes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	addCorsHeaders(w)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(respBytes))
}

func addCorsHeadersForOption(w http.ResponseWriter) {
	addCorsHeaders(w)
	w.Header().Add("Access-Control-Allow-Headers", "Accept,Origin,Content-Type,Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
}
