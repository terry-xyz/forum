APP_NAME=forum
PORT=8080

test:
	go test ./...

run:
	go run .

docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker run --rm -p $(PORT):8080 $(APP_NAME)

docker-run-persist:
	docker run --rm -p $(PORT):8080 -v "$$(pwd)/forum.db:/app/forum.db" $(APP_NAME)

docker: docker-build docker-run

docker-stop-all:
	docker ps -q | xargs -r docker stop