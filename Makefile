APP_NAME := pdf-poster-splitter
APP_ID ?= com.pdfpostersplitter
MACOSX_SDK_PATH ?=
PKG := .
GOBIN := $(shell go env GOPATH)/bin
ICON_PATH := assets/Icon.png

ifneq (,$(wildcard .env))
include .env
endif

.PHONY: help tidy fmt test build build-linux build-darwin build-windows run clean package package-all fyne-install cross-install cross check-fyne

help:
	@echo "Available targets:"
	@echo "  make tidy          - download/update Go dependencies"
	@echo "  make fmt           - format all Go files"
	@echo "  make test          - run tests"
	@echo "  make build-linux   - package linux app with fyne into dist/"
	@echo "  make build-darwin  - package darwin app for amd64, arm64, and universal into dist/"
	@echo "  make build-windows - package windows app with fyne into dist/"
	@echo "  make run           - run the app"
	@echo "  make clean         - remove build artifacts"
	@echo "  make fyne-install  - install fyne CLI tool"
	@echo "  make cross-install - install fyne-cross tool"
	@echo "  make cross         - build binaries for linux/windows/darwin via fyne-cross"
	@echo "  APP_ID=<id>        - override app ID used for packaging"
	@echo "  MACOSX_SDK_PATH    - path to macOS SDK for darwin cross-build"

tidy:
	go mod tidy

fmt:
	gofmt -w *.go

test:
	go test ./...

check-fyne:
	@command -v fyne >/dev/null 2>&1 || (echo "fyne tool not found. Install with: make fyne-install" && echo "If already installed, ensure this directory is in PATH: $(GOBIN)" && exit 1)

fyne-install:
	go install fyne.io/tools/cmd/fyne@latest

build-linux: check-fyne
	rm -rf dist/linux
	mkdir -p dist/linux
	fyne package -os linux -name "$(APP_NAME)" --icon "$(ICON_PATH)"
	mv -f "$(APP_NAME).tar.xz" "dist/linux/$(APP_NAME).tar.xz"

build-darwin: check-fyne
	rm -rf dist/darwin
	mkdir -p dist/darwin

	GOARCH=amd64 fyne package -os darwin -name "$(APP_NAME)" --icon "$(ICON_PATH)"
	mv -f "$(APP_NAME).app" "$(APP_NAME)-amd64.app"
	GOARCH=arm64 fyne package -os darwin -name "$(APP_NAME)" --icon "$(ICON_PATH)"
	mv -f "$(APP_NAME).app" "$(APP_NAME)-arm64.app"

	cp -R "$(APP_NAME)-arm64.app" "$(APP_NAME)-universal.app"
	lipo -create \
		-output "$(APP_NAME)-universal.app/Contents/MacOS/$(APP_NAME)" \
		"$(APP_NAME)-amd64.app/Contents/MacOS/$(APP_NAME)" \
		"$(APP_NAME)-arm64.app/Contents/MacOS/$(APP_NAME)"

	zip -r "dist/darwin/$(APP_NAME)-amd64.zip" "$(APP_NAME)-amd64.app"
	zip -r "dist/darwin/$(APP_NAME)-arm64.zip" "$(APP_NAME)-arm64.app"
	zip -r "dist/darwin/$(APP_NAME)-universal.zip" "$(APP_NAME)-universal.app"
	rm -rf "$(APP_NAME)-amd64.app" "$(APP_NAME)-arm64.app" "$(APP_NAME)-universal.app"

build-windows: check-fyne
	rm -rf dist/windows
	mkdir -p dist/windowsв
	fyne package -os windows -name "$(APP_NAME)" --icon "$(ICON_PATH)"
	mv -f "$(APP_NAME).exe" "dist/windows/$(APP_NAME).exe"

run:
	mkdir -p dist
	go build -o dist/$(APP_NAME) $(PKG)
	./dist/$(APP_NAME)

clean:
	rm -rf dist

cross-install:
	go install github.com/fyne-io/fyne-cross@latest

cross:
	@test -x "$(GOBIN)/fyne-cross" || (echo "fyne-cross not found. Run: make cross-install" && exit 1)
	@command -v docker >/dev/null 2>&1 || (echo "Docker not found. Install Docker: https://docs.docker.com/engine/install/" && exit 1)
	@docker info >/dev/null 2>&1 || (echo "Cannot connect to Docker. Either start the daemon or add yourself to the docker group:" && echo "  sudo usermod -aG docker \$$USER && newgrp docker" && exit 1)
	$(GOBIN)/fyne-cross linux --icon "$(ICON_PATH)" --output "$(APP_NAME)"
	$(GOBIN)/fyne-cross windows -app-id "$(APP_ID)" --icon "$(ICON_PATH)" --output "$(APP_NAME)"
	@if [ -n "$(MACOSX_SDK_PATH)" ]; then \
		$(GOBIN)/fyne-cross darwin -app-id "$(APP_ID)" -macosx-sdk-path "$(MACOSX_SDK_PATH)" --icon "$(ICON_PATH)" --output "$(APP_NAME)"; \
	else \
		echo "Skipping darwin: set MACOSX_SDK_PATH=/path/to/MacOSX*.sdk to enable"; \
	fi
	rm -rf dist/cross
	mv fyne-cross/dist dist/cross
