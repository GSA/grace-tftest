.PHONY: test lint dependencies
default: test

test: lint
	go test -v ./...

lint: Gopkg.toml dependencies
	dep ensure
	golangci-lint run ./...
	gosec ./...

Gopkg.toml:
ifeq (,$(wildcard Gopkg.toml))
	dep init
endif

dependencies: $(GODEP) $(GOLANGCILINT) $(GOSEC)

$(GOLANGCILINT):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

$(GODEP):
	go get -u github.com/golang/dep/cmd/dep

$(GOSEC):
	go get -u github.com/securego/gosec/cmd/gosec