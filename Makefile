.PHONY: build

build:
	GOOS=darwin GOARCH=amd64 go build -ldflags -o bin/s3-echoer-macos .
	GOOS=linux GOARCH=amd64 go build -ldflags -o bin/s3-echoer-linux .