BINARY  := agent-pulse
GOBIN   := $(shell go env GOPATH)/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/SantiagoBobrik/agent-pulse/cmd.version=$(VERSION)

.PHONY: build install test vet clean release-dry

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) .

install:
	go install -ldflags '$(LDFLAGS)' .

test:
	go test ./... -count=1

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

release-dry:
	goreleaser release --snapshot --clean
