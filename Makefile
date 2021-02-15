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
	go build -i -o gomigrate ./cmd/gomigrate

check test tests: ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q $(GO) test  -count=1 -tags $(TAGS) -timeout $(TIMEOUT)s $(ARGS) ./...

test-integration: ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run integration tests
	$Q GOMAXPROCS=3 $(GO) test -v -p 3 -tags=test_integration -timeout $(TIMEOUT)s $(ARGS) ./tests