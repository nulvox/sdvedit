GOROOT ?= /usr/local/go
PORT   ?= 8080
GO     := $(GOROOT)/bin/go
WASM   := site/main.wasm

.PHONY: all build test lint clean serve

all: build

build: $(WASM)

$(WASM): $(shell find cmd internal -name '*.go')
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM) ./cmd/sdvedit/

test:
	$(GO) test ./internal/...

lint:
	$(GO) vet ./internal/... ./cmd/...

clean:
	rm -f $(WASM)

# Serve the site locally for development (requires Python 3)
serve: build
	cd site && python3 -m http.server $(PORT)
