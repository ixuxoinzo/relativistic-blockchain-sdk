.PHONY: build test run docker clean deploy lint security cover bench

BINARY_NAME=relativisticd
DOCKER_IMAGE=ixuxoinzo/relativistic-sdk
VERSION=1.0.0

build:
	@go build -o bin/$(BINARY_NAME) ./cmd/relativisticd

test:
	@go test ./... -v

run: build
	@./bin/$(BINARY_NAME)

docker-build:
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .

clean:
	@rm -rf bin/
	@docker system prune -f

deploy: docker-build
	@docker push $(DOCKER_IMAGE):$(VERSION)

lint:
	@golangci-lint run

all: lint test build