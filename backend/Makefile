.PHONY: build generate test test-unit test-integration test-public-release-regression test-e2e

VERSION ?= $(shell tr -d '\r\n' < ./cmd/server/VERSION)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

generate:
	go generate ./ent
	go generate ./cmd/server

test:
	go test ./...
	golangci-lint run ./...

test-unit:
	go test -tags=unit ./...

test-integration:
	go test -tags=integration ./...

test-public-release-regression:
	go test -tags=unit -run 'TestPublicReleaseAPIRegression|TestSettingHandler_GetPublicSettings_PublicReleaseDashboardContract' ./internal/server/routes ./internal/handler
	go test -tags=integration -run TestPublicReleaseHTTPIntegrationGatewayEntrypoints ./internal/publicrelease

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	go test -tags=e2e -v -timeout=300s ./internal/integration/...
