gofiles := $(shell find . -name '*.go')
goflag :=

all: build

test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.out  ./...
	@echo To view coverage graph use go tool cover -html=coverage.out

golint:
	go fmt ./...
	go vet ./...

build: build-cli

build-cli: bin/feel

bin/feel: ${gofiles}
	go build $(goflag) -o $@ cli/main.go

clean:
	rm -rf build dist bin/*

.PHONY: test gofmt build-cli clean
.SECONDARY: $(buildarchdirs)
