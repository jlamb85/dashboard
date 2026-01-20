# Makefile for the server-dashboard project

.PHONY: all build run clean

all: build

build:
	go build -o server-dashboard main.go

run: build
	./server-dashboard

clean:
	go clean
	rm -f server-dashboard

test:
	go test ./... -v

fmt:
	go fmt ./...