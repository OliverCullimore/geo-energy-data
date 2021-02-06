# Set Variables
FOLDERPATH=$(shell pwd)
FOLDERNAME=$(shell basename "$(PWD)")
GOFILES=$(find . -type f -name '*.go')
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
BINARY_WIN=$(FOLDERNAME).exe
BINARY_UNIX=$(FOLDERNAME)

# Set default binary name based on operating system
ifeq ($(OS), Windows_NT)
	BINARY_NAME=$(BINARY_WIN)
else
	BINARY_NAME=$(BINARY_UNIX)
endif

# Make is verbose. Make it silent.
MAKEFLAGS += --silent

# Define help output
.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## build: Builds the binary.
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

#test:
#	$(GOTEST) -v ./...

# clean: Cleans up.
#clean:
#	$(GOCLEAN)
#	rm -f $(BINARY_NAME)
#	rm -f $(BINARY_UNIX)

## run: Builds then runs the binary.
run:
	$(MAKE) build && ./$(BINARY_NAME)

# deps: Installs any missing dependencies.
#deps:
#	$(GOGET) github.com/markbates/goth
#	$(GOGET) github.com/markbates/pop

## fmt: Formats all .go files found.
fmt:
	$(GOFMT) ./...

# Cross compilation
## build-linux: Builds the binary for linux.
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

# build-docker: Builds a docker image.
#build-docker:
#	docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_UNIX)" -v