APP_NAME=forum
CONTAINER_NAME=forum
PORT=8080
DATABASE_PATH=forum.db
CONTAINER_DB_PATH=/app/data/forum.db
DATA_DIR=data

.PHONY: test test-long run run-config seed docker-build docker-rebuild docker-test docker-test-long docker-run docker-run-persist docker-run-detached docker docker-restart docker-logs docker-stop docker-stop-all docker-clean

test:
	go test ./...

test-long:
	go test -count=10 ./...

run:
	go run .

run-config:
	PORT=$(PORT) DATABASE_PATH=$(DATABASE_PATH) go run .

seed:
	DATABASE_PATH=$(DATABASE_PATH) go run ./cmd/seed

docker-build:
	docker build -t $(APP_NAME) .

docker-rebuild:
	docker build --no-cache -t $(APP_NAME) .

docker-test:
	docker run --rm -v "$$(pwd):/app" -w /app golang:1.24 go test ./...

docker-test-long:
	docker run --rm -v "$$(pwd):/app" -w /app golang:1.24 go test -count=10 ./...

docker-run:
	docker run --rm -p $(PORT):8080 -e PORT=8080 $(APP_NAME)

docker-run-persist:
	mkdir -p $(DATA_DIR)
	docker run --rm -p $(PORT):8080 -e PORT=8080 -e DATABASE_PATH=$(CONTAINER_DB_PATH) -v "$$(pwd)/$(DATA_DIR):/app/data" $(APP_NAME)

docker-run-detached:
	mkdir -p $(DATA_DIR)
	docker run -d --name $(CONTAINER_NAME) -p $(PORT):8080 -e PORT=8080 -e DATABASE_PATH=$(CONTAINER_DB_PATH) -v "$$(pwd)/$(DATA_DIR):/app/data" $(APP_NAME)

docker: docker-build docker-run

docker-restart: docker-stop docker-run-detached

docker-logs:
	docker logs -f $(CONTAINER_NAME)

docker-stop:
	-docker rm -f $(CONTAINER_NAME)

docker-stop-all:
	docker ps -q | xargs -r docker stop

docker-clean:
	-docker rm -f $(CONTAINER_NAME)
	-docker rmi $(APP_NAME)
