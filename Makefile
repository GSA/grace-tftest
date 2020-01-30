GOBIN := $(GOPATH)/bin
GOLANGCILINT := $(GOBIN)/golangci-lint

.PHONY: test lint dependencies
default: test

test: lint
	go test -v ./...

lint: dependencies
	$(GOLANGCILINT) run ./...

dependencies: $(GOLANGCILINT) go.sum

$(GOLANGCILINT):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

go.sum: go.mod
	go mod tidy

go.mod:
	go mod init