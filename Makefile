NAME := prolificli
DOCKER_PREFIX = benmatselby
DOCKER_RELEASE ?= latest

.PHONY: explain
explain:
	### Welcome
	#
	# .______   .______        ______    __       __   _______  __    ______  __       __
	# |   _  \  |   _  \      /  __  \  |  |     |  | |   ____||  |  /      ||  |     |  |
	# |  |_   | |  |_   |    |  |  |  | |  |     |  | |  |__   |  | |  ,----'|  |     |  |
	# |   ___/  |      /     |  |  |  | |  |     |  | |   __|  |  | |  |     |  |     |  |
	# |  |      |  |\  \----.|  `--'  | |  `----.|  | |  |     |  | |  `----.|  `----.|  |
	# | _|      | _| `._____| \______/  |_______||__| |__|     |__|  \______||_______||__|
	#
	#
	### Installation
	#
	# $$ make all
	#
	### Targets
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

GITCOMMIT := $(shell git rev-parse --short HEAD)

.PHONY: clean
clean: ## Clean the local dependencies
	rm -fr vendor

.PHONY: install
install: ## Install the local dependencies
	go install github.com/golang/mock/mockgen@master
	go install github.com/securego/gosec/cmd/gosec@master
	go install golang.org/x/lint/golint@master
	go get ./...

.PHONY: vet
vet: ## Vet the code
	go vet -v ./...

.PHONY: lint
lint: ## Lint the code
	golint -set_exit_status $(shell go list ./... | grep -v vendor)

.PHONY: security
security: ## Inspect the code
	gosec ./...

.PHONY: build
build: ## Build the application
	go build .

.PHONY: static
static: ## Build the application
	CGO_ENABLED=0 go build -ldflags "-extldflags -static -X github.com/benmatselby/$(NAME)/version.GITCOMMIT=$(GITCOMMIT)" -o $(NAME) .

.PHONY: test
test: ## Run the unit tests
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: test-cov
test-cov: test ## Run the unit tests with coverage
	go tool cover -html=coverage.out

.PHONY: test-gen-mock
test-gen-mock: ## Generate the mocks
	@mockgen -source client/client.go > mock_client/mock_client.go

.PHONY: all ## Run everything
all: clean install vet build test

.PHONY: static-all ## Run everything
static-all: clean install vet static test

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_PREFIX)/$(NAME):$(DOCKER_RELEASE) --platform=amd64 .

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_PREFIX)/$(NAME):$(DOCKER_RELEASE)
