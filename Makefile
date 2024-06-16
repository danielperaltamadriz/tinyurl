
.PHONY: precommit
precommit: build-api lint test


# Build

.PHONY: build-api
build-api:
	go build -o bin/api cmd/api/main.go


# Run

.PHONY: run-api
run-api:
	go run cmd/api/main.go

# Test

.PHONY: test
test:
	go test ./...

.PHONY: test-acceptance
test-acceptance:
	go install github.com/onsi/ginkgo/v2/ginkgo &&\
	ginkgo ./acceptance_test/

.PHONY: test-bench
test-bench:
	go test -bench=. ./...


# Lint

.PHONY: lint
lint:
	golangci-lint run


# Docker

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build -t tinyurl .

.PHONY: docker-run-api
docker-run-api:
	docker run -p 8080:8080 -t tinyurl

.PHONY: docker-compose-up
docker-compose-up:
	docker compose up --build
