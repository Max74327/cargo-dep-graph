BIN_DIR   := bin
BIN       := $(BIN_DIR)/cargo-dep-graph
PKG       := ./cmd/cargo-dep-graph
CONFIG    ?= config.yaml
GO        ?= go
GOFLAGS   ?=
LINT      ?= golangci-lint
FMT       ?= gofmt

.PHONY: all build test lint fmt fmt-check vet run clean nix-build nix-shell

all: build

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
		echo "Не отформатировано:"; \
		echo "$$out"; \
		exit 1; \
	fi

vet:
	$(GO) vet $(GOFLAGS) ./...

run: build
	./$(BIN) $(CONFIG)

clean:
	rm -rf $(BIN_DIR)

## Собрать через Nix (если nix доступен).
nix-build:
	nix build

nix-shell:
	nix develop
