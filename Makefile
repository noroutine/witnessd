TAG=v0.0.7

.PHONY: all

all: dep build test

build: dominion

dep:
	go get github.com/Masterminds/glide
	glide install

dominion: dep *.go
	go build

test:
	go test -v ./...

run:
	go run

container: build
	./prepare_docker.sh

push-container: container
	docker tag dominion gcr.io/dominion-p2p/dominion:${TAG}
	./deploy_gcp.sh ${TAG}
	
