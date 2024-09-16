package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	workerThreadStatus = "running"
	timerThreadStatus  = "running"
	logs               = []string{}
	logsMutex          sync.Mutex
	isServiceRunning   = false                  // To track whether the service is running
	logsLabel          *widget.Label            // Global reference to logsLabel in the GUI
	server             *http.Server             // To control the HTTP server lifecycle
	w                  fyne.Window              // Global window reference for updating UI
	commandQueue       = make(chan string, 100) // Command queue for worker thread
)

func main() {

	a := app.New()
	w = a.NewWindow("File Modification Tracker UI")

	logsLabel = widget.NewLabel("Logs will be shown here...")

	//ADD NOTE HERE
	startButton := widget.NewButton("Start Service", func() {
		go startService() // Run startService in a goroutine
		refreshLogs()     // Ensure logsLabel is initialized before this is called
	})
	stopButton := widget.NewButton("Stop Service", func() {
		stopService()
		refreshLogs()
	})
	viewLogsButton := widget.NewButton("View Logs", func() {
		refreshLogs()
	})

	content := container.NewVBox(
		startButton,
		stopButton,
		viewLogsButton,
		logsLabel,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()

}

// ADD NOTE HERE
func startService() {
	if isServiceRunning {
		appendLog("Service is already running")
		refreshLogs()
		return
	}

	config, err := LoadConfig(".")
	if err != nil {
		log.Println("Error loading config: ", err)
		appendLog(fmt.Sprintf("Error loading config: %s", err.Error()))
	}

	isServiceRunning = true
	workerThreadStatus = "running"
	timerThreadStatus = "running"

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", HealthCheckHandler)
	r.Get("/logs", LogsHandler)
	r.Post("/add-command", AddCommandHandler)

	server = &http.Server{
		Addr:    ":" + config.Port,
		Handler: r,
	}

	go func() {
		fmt.Printf("Starting server on port %s...\n", config.Port)
		appendLog(fmt.Sprintf("Server started on port %s", config.Port))
		refreshLogs()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Error starting server: ", err)
			appendLog(fmt.Sprintf("Error starting server: %s", err.Error()))

		}
	}()

	//ADD NOTE HERE

	// Start the TimerThread in a separate goroutine
	go TimerThread(config.Frequency, config.Directory, config.Endpoint)
	go WorkerThread()

	// Log the service start event
	fmt.Println("Service started")
	appendLog("Service started")
	refreshLogs()

	// Signal handling for graceful shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

		// Wait for termination signal
		<-stop
		fmt.Println("Shutting down service...")
		appendLog("Shutting down service...")

		// Stop the service
		stopService()
	}()
}

// Stop the Go binary process
//ADD NOTE HERE

func stopService() {
	if !isServiceRunning {
		fmt.Println("No service running")
		appendLog("No service running")
		return
	}

	// Stop the HTTP server gracefully
	fmt.Println("Stopping server...")
	appendLog("Stopping server...")

	// Create a context with timeout to allow graceful shutdown of the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
		appendLog(fmt.Sprintf("Error shutting down server: %v", err))
	} else {
		fmt.Println("Server stopped")
		appendLog("Server stopped")
	}

	// Set flags to stop the worker and timer threads
	isServiceRunning = false
	workerThreadStatus = "stopped"
	timerThreadStatus = "stopped"

	// Log service stop event
	fmt.Println("Service stopped")
	appendLog("Service stopped")
	refreshLogs()
}
