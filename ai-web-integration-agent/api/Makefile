.PHONY: build run docker docker-run clean

# Build the API server
build:
	go build -o web-integration-api .

# Run the API server
run: build
	./web-integration-api

# Build the Docker image
docker:
	docker build -t web-integration-api -f Dockerfile ..

# Run the Docker container
docker-run: docker
	docker run -p 8080:8080 -v browser-data:/data/browser-data -v screenshots:/data/screenshots -v logs:/data/logs web-integration-api

# Clean up
clean:
	rm -f web-integration-api

