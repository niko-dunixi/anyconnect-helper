.DEFAULT: build
.PHONY: build test clean

ifeq ($(OS),Windows_NT)
    GOOS:=windows
    GOARCH:=x86_64
    ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
        GOARCH:=amd64
    else
        ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
            GOARCH:=amd64
        endif
        ifeq ($(PROCESSOR_ARCHITECTURE),x86)
            GOARCH:=x86_64
        endif
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        GOOS:=linux
        UNAME_P := $(shell uname -p)
        ifeq ($(UNAME_P),x86_64)
            GOARCH:=amd64
        endif
        ifneq ($(filter %86,$(UNAME_P)),)
            GOARCH:=x86_64
        endif
        ifneq ($(filter arm%,$(UNAME_P)),)
            GOARCH:=arm
        endif
    endif
    ifeq ($(UNAME_S),Darwin)
        GOOS:=darwin
        GOARCH:=amd64
    endif
endif

ifndef GOPATH
    GOPATH:=$(HOME)/go
endif
GO111MODULE:=on
GO_VERSION:=1.12
GO:=docker run --rm -v "$(GOPATH):/go" -v "$(PWD)":/usr/src/redowl-connect -w /usr/src/redowl-connect -e GO111MODULE=$(GO111MODULE) -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) golang:$(GO_VERSION) go
# Private doesn't entirely convey the intent of what it's used for.
# When go executes tests, it creates an executable binary based on your
# test sources. So if you're attempting to cross-compile to OSX you'll fail
# your tests because the docker container is linux based and cannot run an
# OSX executable.
GO_PRIVATE:=docker run --rm -v "$(GOPATH):/go" -v "$(PWD)":/usr/src/redowl-connect -w /usr/src/redowl-connect -e GO111MODULE=$(GO111MODULE) golang:$(GO_VERSION) go

build:
	mkdir -p "${GOPATH}"
	$(GO) build

test:
	$(GO_PRIVATE) test

clean:
	git clean -xdf