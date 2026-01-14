GO=go
DIST_DIR=dist
BINARY_NAME=$(shell basename $(PWD))
BINARY_PATH=$(DIST_DIR)/$(BINARY_NAME)

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

run:
	@test -f .env && set -a && . ./.env && set +a; $(GO) run ./cmd/app

# Docker targets
docker-build:
	docker compose -f docker-compose.build.yml build

docker-up:
	docker compose up -d

docker-up-build:
	docker compose -f docker-compose.build.yml up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	@echo "Cleaning up Docker resources..."
	@docker ps -a --filter "name=$(BINARY_NAME)" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
	@docker images --filter "reference=$(BINARY_NAME)*" --format "{{.Repository}}:{{.Tag}}" | xargs -r docker rmi -f 2>/dev/null || true
	@echo "Docker cleanup complete"

docker-clean-all: docker-clean
	@echo "Cleaning up Docker build cache..."
	@docker builder prune -f
	@echo "Docker cleanup complete (including build cache)"

.PHONY: all build build-prod clean format test run \
        docker-build docker-up docker-up-build docker-down docker-logs docker-clean docker-clean-all
