APP_NAME = gomigrate
TAGS ?= ""
APP_VERSION = -X ${PACKAGE}/cmd/command.appName=${APP_NAME} \
	-X ${PACKAGE}/cmd/command.version=${VERSION} \
	-X ${PACKAGE}/cmd/command.branch=${BRANCH} \
	-X ${PACKAGE}/cmd/command.revision=${REVISION} \
	-X ${PACKAGE}/cmd/command.buildDate=${DATE} \
	-X ${PACKAGE}/cmd/command.buildUser=${USER} \
	-X ${PACKAGE}/cmd/command.goVersion=${GOVERSION}

LD_FLAGS = "-w -s $(APP_VERSION)"

GO       = go
GODOC    = godoc
GOFMT    = gofmt
GOIMPORTS    = goimports
DOCKER   = docker
TIMEOUT  = 15
GOLINT = golangci-lint
GOOS	 = $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH	 ?= amd64
CGO_ENABLED ?= 0

Q = $(if $(filter 1,$V),,@)

build:
	$Q $(GO) build -i -o ./bin ./cmd/gomigrate

check test tests: ; $(info $(M) running tests…) @ ## Run tests
	$Q $(GO) test -v -race -count 100 ./...

test-integration: ; $(info $(M) running integration tests…) @ ## Run integration tests
	$Q GOMAXPROCS=6 $(GO) test -count=1 -timeout=30s -v -p 6 -tags=test_integration ./tests

test-integration-docker: ;
	$Q docker-compose -f .docker/docker-compose.yml up --build