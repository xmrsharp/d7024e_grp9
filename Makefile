# DOCKER TASKS
# Terminal: make [command]

# Build the container
build:
	docker build . -t kadlab

# Run the container
run:
	docker compose up

# Run the container go mod tidy && d
rb:
	docker build . -t kadlab && docker compose up

# Down the container
down:
	docker compose down

# Stop and remove a running container
stop: 
	docker stop kadlab; docker rm kadlab

# Returns test coverage
cov:
	go test -v -coverprofile cover.out ./... | grep D7024E
	go tool cover -html cover.out -o cover.html 
	go tool cover -func cover.out | grep -Eo "[0-9]+\.[0-9]+" | tail -n 1
	