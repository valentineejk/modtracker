package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CommandRequest struct {
	Command string `json:"command" validate:"required"`
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"workerThread": workerThreadStatus,
		"timerThread":  timerThreadStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogsHandler returns the collected logs
func LogsHandler(w http.ResponseWriter, r *http.Request) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func AddCommandHandler(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Command == "" {
		http.Error(w, "Command cannot be empty", http.StatusBadRequest)
		return
	}

	// Add the command to the worker thread queue
	AddCommandToQueue(req.Command)

	w.WriteHeader(http.StatusOK)
	appendLog("Command added to queue")
	fmt.Fprintln(w, "Command added to queue")
}
