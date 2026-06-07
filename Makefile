.PHONY: build run validate test fmt vet clean tidy all

BINARY_NAME=sso-testkit
CONFIG?=configs/scenarios/oidc-token-exchange.yaml
MODE?=stub
REPORT?=-
GO?=go
GO111MODULE=on

build:
	GO111MODULE=$(GO111MODULE) $(GO) build -o bin/$(BINARY_NAME) ./cmd/sso-testkit

run:
	GO111MODULE=$(GO111MODULE) $(GO) run ./cmd/sso-testkit run --config $(CONFIG) --mode $(MODE) --report $(REPORT)

validate:
	GO111MODULE=$(GO111MODULE) $(GO) run ./cmd/sso-testkit validate-config --config $(CONFIG) --mode $(MODE)

test:
	GO111MODULE=$(GO111MODULE) $(GO) test ./...

fmt:
	GO111MODULE=$(GO111MODULE) $(GO) fmt ./...

vet:
	GO111MODULE=$(GO111MODULE) $(GO) vet ./...

clean:
	rm -rf bin/
	GO111MODULE=$(GO111MODULE) $(GO) clean

tidy:
	GO111MODULE=$(GO111MODULE) $(GO) mod tidy

all: fmt vet test build
