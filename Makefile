.PHONY: all

all: build test

build: dominion

dominion: dominion.go
	go build

test:
	go test -v ./...