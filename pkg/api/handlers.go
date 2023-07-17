package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// getTaskMetricsHandler is the handler function for the GET /metrics/{task_id} endpoint.
// It retrieves the task metrics for the specified task ID and returns them as JSON.
func getTaskMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	// Get the task ID from the request parameters
	vars := mux.Vars(r)
	taskID := vars["task_id"]

	// Retrieve the task metrics for the specified task ID
	taskMetrics, err := getTaskMetrics(taskID)
	switch err.Error() {
	case "token is expired":
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error": "Clickup Token expired. Get a new one and do the request again"}`)
		return
	case "api key is expired or not valid":
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error": "Api Key is expired or is not valid. Get a new one and do the request again"}`)
		return
	}

	// Marshal the task metrics to JSON
	jsonData, err := json.Marshal(taskMetrics)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response
	_, err = w.Write(jsonData)
	if err != nil {
		log.Println("Error writing response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
