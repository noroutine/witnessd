.PHONY: all

TAG=v0.0.7

all: dep build test

build: dominion

dep:
	go get -t ./...

dominion: *.go
	go build

test:
	go test -v ./...
	# cat sample.dominion | ./dominion

container: build
	./prepare_docker.sh

push-container:
	docker tag dominion gcr.io/dominion-p2p/dominion:${TAG}
	./deploy_gcp.sh ${TAG}
	
