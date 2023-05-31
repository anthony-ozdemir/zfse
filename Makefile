# Makefile Options
BUILD_OUTPUT_PREFIX_DIR ?= ./build/
BUILD_OUTPUT_BASE_NAME ?= zfse

.PHONY: help lint test build release clean

help:
	$(info ZFSE - Zone File Search Engine)
	$(info )
	$(info make lint     : Lint the code before publishing)
	$(info make test     : Run all unit tests)
	$(info make build    : Build a single executable for current system)
	$(info make release  : Build multiple executables for all supported systems)
	$(info make clean    : Remove all build artifacts)
	$(info )

lint:
	golangci-lint run

test:
	go test -v ./...

build:
	go build -o ${BUILD_OUTPUT_PREFIX_DIR}${BUILD_OUTPUT_BASE_NAME} ./cmd/main.go

release:
	GOARCH=amd64 GOOS=linux go build -o ${BUILD_OUTPUT_PREFIX_DIR}${BUILD_OUTPUT_BASE_NAME} ./cmd/main.go
	GOARCH=amd64 GOOS=windows go build -o ${BUILD_OUTPUT_PREFIX_DIR}${BUILD_OUTPUT_BASE_NAME}-win.exe ./cmd/main.go
	GOARCH=amd64 GOOS=darwin go build -o ${BUILD_OUTPUT_PREFIX_DIR}${BUILD_OUTPUT_BASE_NAME}-darwin ./cmd/main.go

clean:
ifeq ($(OS),Windows_NT)
	cmd.exe /C del /Q /F /S "$(BUILD_OUTPUT_PREFIX_DIR)*"
else
	rm -f $(BUILD_OUTPUT_PREFIX_DIR)*
endif