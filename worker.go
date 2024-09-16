package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Worker Thread: Maintains a queue of shell commands and executes them.
// Timer Thread: Sleeps and wakes every minute,
// uses osquery to retrieve file modification stats in the specified directory,
// and logs the results.

// AddCommandToQueue adds a new shell command to the worker queue
func AddCommandToQueue(command string) {
	appendLog(fmt.Sprintf("Adding command to queue: %s", command))
	commandQueue <- command
	refreshLogs()
}

// Executes the shell command and logs the result
func executeCommand(command string) {
	appendLog(fmt.Sprintf("Executing command: %s", command))

	// Check if the command is a control command for the worker service
	controlCommand := strings.ToLower(strings.TrimSpace(command))
	switch controlCommand {
	case "logs":
		refreshLogs()
		return
	case "stop":
		stopService()
		return
	default:
		// If it's not a control command, execute it as a shell command
		cmd := exec.Command("bash", "-c", command)
		output, err := cmd.CombinedOutput()

		if err != nil {
			appendLog(fmt.Sprintf("Error executing command: %v | Output: %s", err, output))
		} else {
			appendLog(fmt.Sprintf("Command executed successfully: %s | Output: %s", command, output))
		}
		refreshLogs()
	}
}

// func executeCommand(command string) {
// 	appendLog(fmt.Sprintf("Executing command: %s", command))

// 	// Check if the command is a control command for the worker service
// 	controlCommand := strings.ToLower(strings.TrimSpace(command))
// 	switch controlCommand {
// 	case "logs":
// 		refreshLogs()
// 		return
// 	case "stop":
// 		stopService()
// 		return
// 	default:
// 		// If it's not a control command, execute it as a shell command
// 		cmd := execCommand("bash", "-c", command)
// 		output, err := cmd.CombinedOutput()

// 		if err != nil {
// 			appendLog(fmt.Sprintf("Error executing command: %v | Output: %s", err, output))
// 		} else {
// 			appendLog(fmt.Sprintf("Command executed successfully: %s | Output: %s", command, output))
// 		}
// 		refreshLogs()
// 	}
// }

// Worker thread -> Executes commands
func WorkerThread() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in WorkerThread:", r)
			appendLog(fmt.Sprintf("Recovered from panic in WorkerThread: %v", r))
			refreshLogs()
		}
	}()

	// Loop until the service is stopped
	for isServiceRunning {
		select {
		case cmd := <-commandQueue:
			executeCommand(cmd)
		default:
			// Sleep briefly if no commands are in the queue to avoid high CPU usage
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("WorkerThread stopped")
	appendLog("WorkerThread stopped")
}
