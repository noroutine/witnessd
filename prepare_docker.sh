#!/bin/bash

cp ./dominion ./docker

(
	cd docker || exit
	docker build -t dominion .
)
