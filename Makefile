GO=go
DIST_DIR=dist
BINARY_NAME=$(shell basename $(PWD))
BINARY_PATH=$(DIST_DIR)/$(BINARY_NAME)

# Targets
all: build

build:
	$(GO) build -o $(BINARY_PATH) ./cmd/app/

build-prod:
	bash scripts/build.sh

clean:
	rm -rf $(DIST_DIR)

format:
	$(GO) fmt ./...

test:
	$(GO) test -v ./...

docker-build:
	docker build -t $(BINARY_NAME):test .

docker-clean:
	@echo "Cleaning up Docker test resources..."
	@docker ps -a --filter "name=$(BINARY_NAME)" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
	@docker images --filter "reference=$(BINARY_NAME)*" --format "{{.Repository}}:{{.Tag}}" | xargs -r docker rmi -f 2>/dev/null || true
	@echo "Docker cleanup complete"

docker-clean-all: docker-clean
	@echo "Cleaning up Docker build cache..."
	@docker builder prune -f
	@echo "Docker cleanup complete (including build cache)"

.PHONY: all build build-prod clean format test docker-build docker-clean docker-clean-all
