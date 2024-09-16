package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var execCommand = exec.Command

var mockExecCommand = func(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, `[{"path":"fakefile1.txt"},{"path":"fakefile2.log"}]`)
	os.Exit(0)
}

func TestRunOsquery(t *testing.T) {
	assert := assert.New(t)

	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	_, err := RunOsquery("/fake/directory")
	assert.NoError(err, "Did not expect an error during osquery execution")

	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	_, err = RunOsquery("/invalid/path")
	assert.NoError(err, "Did not expect an error for invalid path")
}

func TestSendDataToAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	err := SendDataToAPI(ts.URL, []string{"file1.txt", "file2.txt"}, "System stats placeholder")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Simulate API failure
	}))
	defer tsFail.Close()

	err = SendDataToAPI(tsFail.URL, []string{"file1.txt", "file2.txt"}, "System stats placeholder")
	if err == nil {
		t.Errorf("Expected error due to bad request, got nil")
	}
}

func TestExecuteCommand(t *testing.T) {
	assert := assert.New(t)
	execCommand = mockExecCommand                 // Mock the command
	defer func() { execCommand = exec.Command }() // Restore after the test

	// Test valid command
	executeCommand("echo hello")
	assert.Contains(logs[len(logs)-1], "Command executed successfully: echo hello", "Expected successful command execution log")

	// Test invalid command
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false") // Force failure
	}
	executeCommand("invalid")
	assert.Contains(logs[len(logs)-1], "Error executing command", "Expected error log for invalid command")
}

func TestWorkerThread(t *testing.T) {
	logs = []string{}
	isServiceRunning = true
	commandQueue = make(chan string, 2)

	// Start the worker thread in a goroutine
	go WorkerThread()

	// Add two commands to the queue
	AddCommandToQueue("echo hello")
	AddCommandToQueue("echo world")

	time.Sleep(2 * time.Second) // Give the worker time to process the commands

	// Use helper function to check for log content
	assertLogContains(t, "Executing command: echo hello")
	assertLogContains(t, "Executing command: echo world")

	isServiceRunning = false // Stop the worker thread
}

// Helper function to assert log contains specific text
func assertLogContains(t *testing.T, expected string) {
	found := false
	for _, log := range logs {
		if strings.Contains(log, expected) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected log to contain: %s", expected)
	}
}

func TestTimerThread(t *testing.T) {
	// assert := assert.New(t)
	isServiceRunning = true
	execCommand = mockExecCommand                 // Mock the osquery command
	defer func() { execCommand = exec.Command }() // Restore after the test

	// Run the timer thread in the background
	go TimerThread(1, "/Users/test", "http://example.com/api")

	time.Sleep(2 * time.Second) // Allow the timer to run once

	// Check if the logs contain the expected message
	assertLogContains(t, "Checking for file modifications")
	assertLogContains(t, "Error sending data to API: failed to send data to API, status code: 405")

	isServiceRunning = false // Stop the timer thread
}

func TestAppendLog(t *testing.T) {
	initialLogCount := len(logs)
	appendLog("Test log message")

	if len(logs) != initialLogCount+1 {
		t.Errorf("Expected log count to increase by 1, got %d", len(logs))
	}

	if logs[len(logs)-1] != fmt.Sprintf("%s: %s", time.Now().Format("2006-01-02 15:04:05"), "Test log message") {
		t.Errorf("Expected last log message to be 'Test log message', got %s", logs[len(logs)-1])
	}
}

func TestLoadConfig(t *testing.T) {
	// Use assert to simplify test assertions
	assert := assert.New(t)

	// Test successful config load from embedded resource
	config, err := LoadConfig(".")
	assert.NoError(err, "Expected no error while loading config")

	// Check that the config values are correctly parsed
	assert.Equal("3000", config.Port, "Expected port to be 3000")
	assert.Equal(1, config.Frequency, "Expected frequency to be 1")
	assert.Equal("/Users/macbookpro/Documents/savana", config.Directory, "Expected directory to match")
	assert.Equal("https://google.com", config.Endpoint, "Expected endpoint to match")

	// Test when resourceAppEnv is nil (simulate missing embedded resource)
	resourceAppEnv = nil
	_, err = LoadConfig(".")
	assert.Error(err, "Expected error for missing embedded resource")
	assert.Contains(err.Error(), "no embedded resource found", "Error message should indicate missing resource")
}
