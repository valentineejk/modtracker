
#build the go binary
mac:
	go build

#build pkg for Mac with installation
mactar:
	# fyne bundle -o bundled.go app.env
	fyne package -os darwin -icon icon.png

# WINDOWS BINARY
windows:
	GOOS="windows" go build

# WINDOWS EXECUTABLE
windowsexe:
	fyne package -os windows -icon icon.png

# RUN THE BINARY
run:
	./modtracker

#TEST THE PROJECT
test:
	go test ./...