NAME := prolific
DOCKER_PREFIX = prolific-oss
DOCKER_RELEASE ?= latest
GIT_ORG = prolific-oss
BUILD_DIR ?= build
GOOS ?=
ARCH ?=
OUT_PATH=$(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH)
GIT_RELEASE ?= $(shell git rev-parse --short HEAD)

.PHONY: explain
explain:
	### Welcome
	#
	# .______   .______        ______    __       __   _______  __    ______
	# |   _  \  |   _  \      /  __  \  |  |     |  | |   ____||  |  /      |
	# |  |_   | |  |_   |    |  |  |  | |  |     |  | |  |__   |  | |  ,----'
	# |   ___/  |      /     |  |  |  | |  |     |  | |   __|  |  | |  |
	# |  |      |  |\  \----.|  `--'  | |  `----.|  | |  |     |  | |  `----.
	# | _|      | _| `._____| \______/  |_______||__| |__|     |__|  \______|
	#
	#
	### Installation
	#
	# $$ make all                      # Build and test locally
	# $$ make install-binary           # Install to $$GOPATH/bin
	# $$ go install ./cmd/prolific     # Alternative: install directly
	#
	### Targets
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Clean build artifacts and dependencies
	rm -fr vendor
	rm -fr $(BUILD_DIR) && mkdir -p $(BUILD_DIR)
	rm -f $(NAME)

.PHONY: install
install: install-binary ## Install dependencies and the prolific binary
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	go install github.com/golang/mock/mockgen@master
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go get ./...

.PHONY: install-binary
install-binary: ## Install the prolific binary to $GOPATH/bin
	go install ./cmd/prolific

.PHONY: lint
lint: ## Vet the code
	golangci-lint run

.PHONY: lint-dockerfile
lint-dockerfile: ## Lint the dockerfile
	npx dockerfilelint Dockerfile

.PHONY: build
build: ## Build the application
	go build -o $(NAME) ./cmd/prolific

.PHONY: static
static: ## Build the application
	CGO_ENABLED=0 go build \
		-ldflags "-extldflags -static -X github.com/$(GIT_ORG)/$(NAME)/version.GITCOMMIT=$(GIT_RELEASE)" \
		-o $(NAME) ./cmd/prolific

.PHONY: static-named
static-named: ## Build the application with named outputs
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 \
		go build \
		-ldflags "-extldflags -static -X github.com/$(GIT_ORG)/$(NAME)/version.GITCOMMIT=$(GIT_RELEASE)" \
		-o $(OUT_PATH) ./cmd/prolific

	md5sum $(OUT_PATH) > $(OUT_PATH).md5 || md5 $(OUT_PATH) > $(OUT_PATH).md5
	sha256sum $(OUT_PATH) > $(OUT_PATH).sha256 || shasum $(OUT_PATH) > $(OUT_PATH).sha256

.PHONY: test
test: ## Run the unit tests
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: test-cov
test-cov: test ## Run the unit tests with coverage
	go tool cover -html=coverage.out -o build/coverage.html

.PHONY: test-gen-mock
test-gen-mock: ## Generate the mocks
	@mockgen -source client/client.go > mock_client/mock_client.go

.PHONY: all ## Run everything
all: clean install build test

.PHONY: static-all ## Run everything
static-all: clean install static test

.PHONY: docker-build
docker-build: ## Build the docker image
	docker build -t $(DOCKER_PREFIX)/$(NAME):$(DOCKER_RELEASE) .

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(DOCKER_PREFIX)/$(NAME):$(DOCKER_RELEASE)

.PHONY: docker-scout
docker-scout: ## Check the Docker image for vulnerabilities
	docker scout cves $(DOCKER_PREFIX)/$(NAME):$(DOCKER_RELEASE)

.PHONY: format-changed
format-changed: ## Format only changed Go files
	@git diff --name-only HEAD | grep '\.go$$' | xargs -r gofmt -w
	@git diff --name-only HEAD | grep '\.go$$' | xargs -r goimports -w