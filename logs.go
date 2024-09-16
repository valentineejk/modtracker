package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Append a log message to the logs and protect with mutex
func appendLog(message interface{}) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	var logEntry string

	switch msg := message.(type) {
	case string:
		// format based on str
		logEntry = fmt.Sprintf("%s: %s", time.Now().Format("2006-01-02 15:04:05"), msg)
	case []byte:
		// format based on json
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal(msg, &prettyJSON); err != nil {
			// not jsin
			logEntry = fmt.Sprintf("%s: %s", time.Now().Format("2006-01-02 15:04:05"), string(msg))
		} else {
			// convert to json
			prettyBytes, err := json.MarshalIndent(prettyJSON, "", "  ")
			if err != nil {
				logEntry = fmt.Sprintf("%s: %s", time.Now().Format("2006-01-02 15:04:05"), string(msg))
			} else {
				logEntry = fmt.Sprintf("%s: %s", time.Now().Format("2006-01-02 15:04:05"), prettyBytes)
			}
		}
	default:
		// For any other types, log them as a generic string
		logEntry = fmt.Sprintf("%s: %v", time.Now().Format("2006-01-02 15:04:05"), msg)
	}

	// Append the log entry to the logs
	logs = append(logs, logEntry)

	// Optionally log the entry to the console (for debugging)
	log.Println(logEntry)
}

// Refresh the logs in the Fyne UI
func refreshLogs() {
	if logsLabel == nil {
		fmt.Println("logsLabel is nil!")
		return
	}

	if w == nil {
		fmt.Println("Window (w) is nil! Cannot refresh window content.")
		return
	}

	logsMutex.Lock()
	defer logsMutex.Unlock()

	if len(logs) == 0 {
		logsLabel.SetText("No logs available")
	} else {
		logsText := ""
		for _, logEntry := range logs {
			logsText += logEntry + "\n"
		}
		logsLabel.SetText(logsText)
	}
	w.Content().Refresh()
}
