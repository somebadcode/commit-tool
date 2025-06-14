.PHONY: build prerequisites test check mod-check mod-download mod-tool-download mod-tool

GOOSS = linux darwin windows
ARCHS = amd64 arm64

build: check test build-linux-amd64 build-linux-arm64 build-darwin-arm64 build-darwin-amd64 build-windows-amd64 build-windows-arm64

prerequisites:
	@if ! command -v go >/dev/null; then echo "Go is required! ( https://go.dev/dl/ )" && exit 1; fi

test:
	@echo Running unit tests...
	@go test -cover -coverprofile=profile.cov -covermode=atomic ./...
	@go tool cover -func profile.cov

check: prerequisites mod-check mod-download
	@echo Vetting code...
	@go vet ./...

	@echo Static check...
	@go tool -modfile=go.tool.mod staticcheck

	@echo Vulnerability check...
	@go tool -modfile=go.tool.mod govulncheck

mod-check:
	@echo Verifying mod file!
	@go mod verify

	@echo Verifying tool mod file!
	@go mod verify -modfile=go.tool.mod

mod-download:
	@echo Downloading modules...
	@go mod download

mod-tool:
	@echo Installing tooling...
	@go get -u -tool -modfile=go.tool.mod honnef.co/go/tools/cmd/staticcheck
	@go get -u -tool -modfile=go.tool.mod golang.org/x/vuln/cmd/govulncheck

clean:
	@rm -rf dist/

define OS_ARCH_template
.PHONY build-$(1)-$(2):
build-$(1)-$(2): dist/$(1)/$(2)/commit-tool$$(if $(filter $(1),windows),.exe)

dist/$(1)/$(2)/commit-tool$$(if $(filter $(1),windows),.exe):
	@echo Building for $(1)/$(2) [output=$$@]...
	@mkdir --parents $$(@D)/
	@GOOS="$(1)" GOARCH="$(2)" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $$@ .
endef

$(foreach arch,$(ARCHS),$(eval $(foreach goos,$(GOOSS), $(eval $(call OS_ARCH_template,$(goos),$(arch)))) ))
