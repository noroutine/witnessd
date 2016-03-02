cp ./dominion ./docker
(
	cd docker
	docker build -t dominion .
)

rm ./docker/dominion
