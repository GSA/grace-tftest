GOBIN := $(GOPATH)/bin
GOLANGCILINT := $(GOBIN)/golangci-lint

.PHONY: test lint dependencies precommit
default: test

test: lint
	go test -v ./...

lint: dependencies
	$(GOLANGCILINT) run ./...

Gopkg.toml: $(GODEP)
ifeq (,$(wildcard Gopkg.toml))
	$(GODEP) init
endif

dependencies: $(GOLANGCILINT) $(GOSEC) Gopkg.toml precommit

$(GOLANGCILINT):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

go.sum: go.mod
	go mod tidy

$(GOSEC):
	go get -u github.com/securego/gosec/cmd/gosec

precommit:
ifneq ($(strip $(hooksPath)),.github/hooks)
	@git config --add core.hooksPath .github/hooks
endif
