GOBIN := $(GOPATH)/bin
GODEP := $(GOBIN)/dep
GOLANGCILINT := $(GOBIN)/golangci-lint
GOSEC := $(GOBIN)/gosec

.PHONY: test lint dependencies precommit
default: test

test: lint
	go test -v ./...

lint: dependencies
	$(GODEP) ensure
	$(GOLANGCILINT) run ./...
	$(GOSEC) ./...

Gopkg.toml: $(GODEP)
ifeq (,$(wildcard Gopkg.toml))
	$(GODEP) init
endif

dependencies: $(GOLANGCILINT) $(GOSEC) Gopkg.toml precommit

$(GOLANGCILINT):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

$(GODEP):
	go get -u github.com/golang/dep/cmd/dep

$(GOSEC):
	go get -u github.com/securego/gosec/cmd/gosec

precommit:
ifneq ($(strip $(hooksPath)),.github/hooks)
	@git config --add core.hooksPath .github/hooks
endif
