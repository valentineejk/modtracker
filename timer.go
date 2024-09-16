package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func TimerThread(frequency int, directory string, endpoint string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in TimerThread:", r)
			appendLog(fmt.Sprintf("Recovered from panic in TimerThread: %v", r))
			refreshLogs()
		}
	}()

	for isServiceRunning {
		fmt.Println("Checking for file modifications...")
		appendLog(fmt.Sprintln("Checking for file modifications..."))
		refreshLogs()

		// Run osquery to check for file modifications
		files, err := RunOsquery(directory)
		refreshLogs()

		if err != nil {
			fmt.Println("Error running osquery:", err)
			appendLog(fmt.Sprintf("Error running osquery: %v", err))
			refreshLogs()

		} else {
			// Send the data to the API
			err = SendDataToAPI(endpoint, files, "System stats placeholder")
			if err != nil {
				fmt.Println("Error sending data to API:", err)
				appendLog(fmt.Sprintf("Error sending data to API: %v", err))
			} else {
				appendLog("Data collected!")
			}
		}

		// Sleep for the specified frequency before checking again
		time.Sleep(time.Duration(frequency) * time.Minute)
	}

	fmt.Println("TimerThread stopped")
	appendLog("TimerThread stopped")
	refreshLogs()
}

func RunOsquery(directory string) ([]string, error) {
	// Ensure the directory path is properly formatted
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	// SQL query to find files in the specified directory
	query := fmt.Sprintf(
		`SELECT path FROM file WHERE path LIKE '%s%%';`,
		directory,
	)

	// Log the query for debugging
	log.Printf("Running osquery with query: %s", query)

	// Run the osquery command with the constructed query
	cmd := exec.Command("osqueryi", "--json", query)

	// Log the full command being executed for debugging
	log.Printf("Executing osquery command: %v", cmd.Args)

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running osquery: %v", err)
		appendLog(fmt.Sprintf("Error running osquery: %v", err))
		return nil, fmt.Errorf("osquery execution failed: %v", err)
	}

	// Log the raw output from osquery for debugging
	log.Printf("osquery output: %s", output)

	// Parse the JSON output from osquery
	var result []map[string]string
	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Error parsing osquery output: %v", err)
		appendLog(fmt.Sprintf("Error parsing osquery output: %v", err))
		return nil, fmt.Errorf("failed to parse osquery output: %v", err)
	}

	// Collect file paths from the osquery result
	var files []string
	for _, item := range result {
		if path, ok := item["path"]; ok {
			files = append(files, path)
		}
	}

	// Log the files in JSON format
	filesJSON, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		log.Printf("Error converting files to JSON: %v", err)
		appendLog(fmt.Sprintf("Error converting files to JSON: %v", err))
		return nil, fmt.Errorf("failed to convert files to JSON: %v", err)
	}
	appendLog(filesJSON)

	return files, nil
}
