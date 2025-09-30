GOLANGCILINT=golangci-lint
MODERNIZE=go tool modernize
PKGSITE=go tool pkgsite

GO_TEST_FLAGS:=-timeout 5s
ifeq ($(V),1)
GO_TEST_FLAGS:=$(GO_TEST_FLAGS) -v
else
endif
ifeq ($(TEST_RACE),1)
GO_TEST_FLAGS:=$(GO_TEST_FLAGS) -race
endif
ifeq ($(TEST_COVERAGE),1)
GO_TEST_FLAGS:=$(GO_TEST_FLAGS) -coverprofile=coverage.out
endif
ifeq ($(TEST_NO_CACHE),1)
GO_TEST_FLAGS:=$(GO_TEST_FLAGS) -count 1
endif
TEST_PACKAGE ?= ./...
ifneq ($(TEST_RUN),)
GO_TEST_FLAGS := $(GO_TEST_FLAGS) -run $(CHECK_RUN)
endif

.PHONY: all
all:

.PHONY: lint
lint: lint-go

.PHONY: lint-go
lint-go:
	go vet ./...
ifneq ($(CI),1)
	$(GOLANGCILINT) run
endif

.PHONY: lint-go-modernize
lint-go-modernize:
	$(MODERNIZE) -test ./...

.PHON: test
test:
	go test $(GO_TEST_FLAGS) ./...
ifeq ($(COVERAGE),1)
	go tool cover -html=coverage.out -o coverage.html
endif

.PHONY: doc
doc:
	$(PKGSITE) -http localhost:6060 -open
