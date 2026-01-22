APP=inventory-back
PKG=./...
BIN=bin/api
VERSION ?= dev
IMAGE ?= your-registry/$(APP):$(VERSION)

GOLANGCI_LINT_VERSION ?= v1.64.5
MIGRATE ?= migrate
DB_URL ?= $(DATABASE_URL)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt $(PKG)

.PHONY: vet
vet:
	go vet $(PKG)

.PHONY: lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not found"; exit 1)
	golangci-lint run

.PHONY: lint-install
lint-install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: test
test:
	go test $(PKG)

.PHONY: build
build:
	go build -o $(BIN) ./cmd/api

.PHONY: run
run:
	HTTP_ADDR=:8080 go run ./cmd/api

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) .

.PHONY: docker-push
docker-push:
	docker push $(IMAGE)

.PHONY: migrate-up
migrate-up:
	$(MIGRATE) -path ./migrations -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	$(MIGRATE) -path ./migrations -database "$(DB_URL)" down 1

.PHONY: check
check: fmt vet lint test
