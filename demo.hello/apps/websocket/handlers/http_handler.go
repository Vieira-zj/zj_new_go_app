package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"demo.hello/utils"
)

// IndexHandler .
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world")
}

// InitJobResultsHandler .
func InitJobResultsHandler(w http.ResponseWriter, r *http.Request) {
	mock.buildJobResults(10)
	fmt.Fprint(w, "job results init")
}

// GetAllJobResultsHandler .
func GetAllJobResultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		utils.AddCorsHeadersForOptions(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	resultBytes, err := json.Marshal(jobResults)
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

	utils.AddCorsHeadersForSimple(w)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(respBytes))
}
