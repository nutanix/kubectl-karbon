SOURCES := $(shell find . -name '*.go')
PKG := $(shell go list)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
BINARY := kubectl-karbon

BUILD=`date +%FT%T%z`
PLATFORM=`uname`

LDFLAGS=-ldflags "-w -s -X github.com/nutanix/kubectl-karbon/version.Date=${BUILD} -X github.com/nutanix/kubectl-karbon/version.BuiltBy=Makefile"

build: kubectl-karbon

test: $(SOURCES)
	go test -v -short -race -timeout 30s ./...

clean:
	@rm -rf $(BINARY)

clean-install:
	@rm -rf /usr/local/bin/$(BINARY)

install: $(BINARY)
	cp $(BINARY) /usr/local/bin/$(BINARY)

check: ## Static Check Golang files
	@staticcheck ./...

vet: ## go vet files
	@go vet ./...

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o $(BINARY) -${LDFLAGS} ./$(BINARY).go