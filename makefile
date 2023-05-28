ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# Common - can be overridden by .env file or when running make
VERSION ?= 0.0.1
BUILD_INFO ?= Local and manual build

# Override these if building your own images
IMAGE_REG ?= ghcr.io
IMAGE_NAME ?= benc-uk/nanoproxy
IMAGE_TAG ?= latest

# Things you don't want to change
REPO_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

# Tools installed locally into repo, don't change
GOLINT_PATH := $(REPO_DIR)/.tools/golangci-lint
AIR_PATH := $(REPO_DIR)/.tools/air

.EXPORT_ALL_VARIABLES:
.PHONY: help images push lint lint-fix install-tools run build
.DEFAULT_GOAL := help

help: ## 💬 This help message :)
	@figlet $@ || true
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## 🔮 Install dev tools into project bin directory
	@figlet $@ || true
	@$(GOLINT_PATH) > /dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./.tools
	@$(AIR_PATH) -v > /dev/null 2>&1 || ( wget https://github.com/cosmtrek/air/releases/download/v1.42.0/air_1.42.0_linux_amd64 -q -O .tools/air && chmod +x .tools/air )
	
lint: ## 🔍 Lint & format check only, sets exit code on error for CI
	@figlet $@ || true
	$(GOLINT_PATH) run

lint-fix: ## 📝 Lint & format, attempts to fix errors & modify code
	@figlet $@ || true
	$(GOLINT_PATH) run --fix

build: ## 🔨 Build binary into ./bin/ directory
	@figlet $@ || true
	@mkdir -p bin
	@go build -o bin/nanoproxy ./proxy

images: ## 📦 Build container images
	@figlet $@ || true
	docker build . -f build/Dockerfile.proxy -t $(IMAGE_REG)/$(IMAGE_NAME)-proxy:$(IMAGE_TAG) --build-arg VERSION=$(VERSION)
	docker build . -f build/Dockerfile.ingress-ctrl -t $(IMAGE_REG)/$(IMAGE_NAME)-ingress-ctrl:$(IMAGE_TAG) --build-arg VERSION=$(VERSION)

push: ## 📤 Push container images
	@figlet $@ || true
	docker push $(IMAGE_REG)/$(IMAGE_NAME)-proxy:$(IMAGE_TAG)
	docker push $(IMAGE_REG)/$(IMAGE_NAME)-ingress-ctrl:$(IMAGE_TAG)

run-proxy: ## 🎯 Run proxy locally with hot-reload
	@figlet $@ || true
	@$(AIR_PATH) -c proxy/.air.toml

run-ctrl: ## 🎯 Run controller locally with hot-reload
	@figlet $@ || true
	@$(AIR_PATH) -c ingress-ctrl/.air.toml

test: ## 🧪 Run all unit tests
	@figlet $@ || true
	@echo "Not implemented yet"

clean: ## 🧹 Clean up, remove dev data and files
	@figlet $@ || true
	@rm -rf bin
	@rm -rf .tools
	@rm -rf tmp