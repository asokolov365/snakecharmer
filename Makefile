SHELL = bash


GO_MODULES := $(shell find . -name go.mod -exec dirname {} \; | sort)

GOTAGS ?=
GOPATH=$(shell go env GOPATH)
GOARCH?=$(shell go env GOARCH)
MAIN_GOPATH=$(shell go env GOPATH | cut -d: -f1)

export PATH := $(PWD)/bin:$(GOPATH)/bin:$(PATH)

ifeq (, $(shell which golangci-lint))
$(warning "unable to find golangci-lint in $(PATH), run: curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh")
endif

default: all

.PHONY: all
all: fmt test ## Command running by default

.PHONY: go-mod-tidy
go-mod-tidy: $(foreach mod,$(GO_MODULES),go-mod-tidy/$(mod)) ## Run go mod tidy in every module

.PHONY: mod-tidy/%
go-mod-tidy/%:
	@echo "--> Running go mod tidy ($*)"
	@cd $* && go mod tidy


##@ Checks

.PHONY: fmt
fmt: $(foreach mod,$(GO_MODULES),fmt/$(mod)) ## Format go modules

.PHONY: fmt/%
fmt/%:
	@echo "--> Running go fmt ($*)"
	@cd $* && gofmt -s -l -w .

.PHONY: lint
lint: $(foreach mod,$(GO_MODULES),lint/$(mod)) ## Lint go modules and test deps

.PHONY: lint/%
lint/%:
	@echo "--> Running golangci-lint ($*)"
	@cd $* && GOWORK=off golangci-lint run --build-tags '$(GOTAGS)'
	@echo "--> Running enumcover ($*)"
	@cd $* && GOWORK=off enumcover ./...


##@ Testing

.PHONY: cover
cover: ## Run tests and generate coverage report
	go test -tags '$(GOTAGS)' ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: test
test: lint test-all

.PHONY: test-all
test-all: lint $(foreach mod,$(GO_MODULES),test-module/$(mod)) ## Test all

.PHONY: test-module/%
test-module/%:
	@echo "--> Running go test ($*)"
	cd $* && go test $(GOTEST_FLAGS) -tags '$(GOTAGS)' ./...

.PHONY: test-race
test-race: ## Test race
	$(MAKE) GOTEST_FLAGS=-race


##@ Tools

.PHONY: deps
deps: ## Installs Go dependencies.
	@echo "--> Running go get -v"
	go get -v ./...

print-%  : ; @echo $($*) ## utility to echo a makefile variable (i.e. 'make print-GOPATH')

.PHONY: module-versions
module-versions: ## Print a list of modules which can be updated. Columns are: module current_version date_of_current_version latest_version
	@go list -m -u -f '{{if .Update}} {{printf "%-50v %-40s" .Path .Version}} {{with .Time}} {{ .Format "2006-01-02" -}} {{else}} {{printf "%9s" ""}} {{end}}   {{ .Update.Version}} {{end}}' all


##@ Cleanup

.PHONY: clean
clean: ## Removes produced binaries, libraries, and temp files
	@rm -rf $(BIN)
	@rm -f test.log exit-code coverage.out


##@ Help

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
