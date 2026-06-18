.PHONY: build run clean install

# Build the application
build:
	go build -o d4s .

# Run the application
run: build
	./d4s

# Clean build artifacts
clean:
	rm -f d4s

# Install the application to /usr/local/bin
install: build
	sudo mv d4s /usr/local/bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run ./...

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/d4s-linux-amd64 .
	GOOS=linux GOARCH=386 go build -o dist/d4s-linux-x86 .
	GOOS=linux GOARCH=arm GOARM=6 go build -o dist/d4s-linux-armv6 .
	GOOS=linux GOARCH=arm GOARM=7 go build -o dist/d4s-linux-armv7 .
	GOOS=darwin GOARCH=amd64 go build -o dist/d4s-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/d4s-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/d4s-windows-amd64.exe .

