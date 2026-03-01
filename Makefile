BINARY  := agent-pulse
GOBIN   := $(shell go env GOPATH)/bin

.PHONY: build install test vet clean

build:
	go build -o $(BINARY) .

install:
	go install .

test:
	go test ./... -count=1

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
