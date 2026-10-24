BIN_DIR    := bin
BIN        := $(BIN_DIR)/cargo-dep-graph
PKG        := ./cmd/cargo-dep-graph
CONFIG     ?= config.yaml

GO         ?= go
GOFLAGS    ?=
LINT       ?= golangci-lint
FMT        ?= gofmt

.PHONY: all build test lint fmt fmt-check vet check run clean

all: build test run

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN) $(PKG)

test:
	$(GO) test $(GOFLAGS) ./...

lint:
	$(LINT) run

fmt:
	$(FMT) -w .

fmt-check:
	@out="$$($(FMT) -l .)"; \
	if [ -n "$$out" ]; then \
		echo "Not formatted:"; \
		echo "$$out"; \
		exit 1; \
	fi

vet:
	$(GO) vet $(GOFLAGS) ./...

check: fmt-check vet lint test

run: build
	./$(BIN) $(CONFIG)

clean:
	rm -rf $(BIN_DIR)
