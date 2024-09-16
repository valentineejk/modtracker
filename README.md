hi, kindly check the make file to run the project and all the deliverables

This project uses Viper for configuration management. The configuration is stored in the app.env file, which needs to be bundled for the macOS version.

<!-- The app.env file includes settings such as:

	•	PORT: Defines the server port.
	•	FREQUENCY: Sets the interval for the timer thread.
	•	ENDPOINT: API endpoint for sending data.
	•	DIRECTORY: Directory to be monitored. -->


<!-- The application integrates with osquery for system monitoring and file modification tracking. Make sure osquery is installed on the system where this app is being used.

For macOS and Windows:

	•	Install osquery from osquery.io.
	•	The app uses osqueryi to collect file modification stats. -->


<!-- install packages -->
go get fyne.io/fyne/v2
go get github.com/go-chi/chi/v5

<!-- To build the binary for macOS: -->
go build

<!-- To build the binary for Windows: -->
GOOS="windows" GOARCH="amd64" go build -o modtracker.exe

<!-- To package the application for macOS (using Fyne): -->
fyne bundle -o bundled.go app.env

<!-- mac -->
fyne package -os darwin -icon icon.png

<!-- windows -->
fyne package -os windows -icon icon.png


<!-- To run the test cases in the project: -->
go test ./...