.PHONY: build

build:
	go build -o telegram-downloader-bot main.go

build-amd64:
	GOOS=linux GOARCH=amd64 go build -o telegram-downloader-bot main.go
