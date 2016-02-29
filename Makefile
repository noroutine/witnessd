.PHONY: all

all: dep build test

build: dominion

dep:
	go get -t ./...

dominion: *.go
	go build

test:
	go test -v ./...
	cat sample.dominion | ./dominion