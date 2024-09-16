package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Payload struct for the API request
type Payload struct {
	Timestamp     string   `json:"timestamp"`
	ModifiedFiles []string `json:"modified_files"`
	SystemStats   string   `json:"system_stats"`
}

// SendDataToAPI: Sends the collected data to the remote API
func SendDataToAPI(apiEndpoint string, modifiedFiles []string, systemStats string) error {
	//payload
	payload := Payload{
		Timestamp:     time.Now().Format(time.RFC3339),
		ModifiedFiles: modifiedFiles,
		SystemStats:   systemStats,
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send the JSON payload to the remote API
	resp, err := http.Post(apiEndpoint, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send data to API, status code: %d", resp.StatusCode)
	}

	fmt.Println("Successfully sent data to API")
	return nil
}
