# Set ENV to dev, prod, etc. to load .env.$(ENV) file
ENV ?= 
-include .env
export
-include .env.$(ENV)
export

# Internal variables you don't want to change
SHELL := /bin/bash
MAKEFLAGS += --warn-undefined-variables --no-builtin-rules
REPO_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GOLINT_PATH := $(REPO_DIR)/.tools/golangci-lint
AIR_PATH := $(REPO_DIR)/.tools/air

.EXPORT_ALL_VARIABLES:
.PHONY: help images push lint lint-fix install-tools run-proxy run-ctrl release test build check-vars
.DEFAULT_GOAL := help

print-env: ## 🚿 Print all env vars for debugging
	@figlet $@ || true
	@echo "Environment: ${ENV}"
	@echo "VERSION: $(VERSION)"
	@echo "IMAGE_REG: $(IMAGE_REG)"
	@echo "IMAGE_NAME: $(IMAGE_NAME)"
	@echo "IMAGE_TAG: $(IMAGE_TAG)"

help: ## 💬 This help message :)
	@figlet $@ || true
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## 🔮 Install dev tools into project bin directory
	@figlet $@ || true
	@$(GOLINT_PATH) > /dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./.tools
	@$(AIR_PATH) -v > /dev/null 2>&1 || ( wget https://github.com/cosmtrek/air/releases/download/v1.42.0/air_1.42.0_linux_amd64 -q -O .tools/air && chmod +x .tools/air )
	
lint: ## 🔍 Lint & format check only, sets exit code on error for CI
	@figlet $@ || true
	$(GOLINT_PATH) run --timeout 3m
	npx prettier --check . '!deploy/helm/nanoproxy/templates/**'

lint-fix: ## 📝 Lint & format, attempts to fix errors & modify code
	@figlet $@ || true
	$(GOLINT_PATH) run --fix
	npx prettier --write . '!deploy/helm/nanoproxy/templates/**'

build: ## 🔨 Build binary into ./bin/ directory
	@figlet $@ || true
	@mkdir -p bin
	@go build -o bin/nanoproxy ./proxy
	@go build -o bin/controller ./controller

images: check-vars ## 📦 Build container images
	@figlet $@ || true
	docker build . -f build/Dockerfile.proxy -t $(IMAGE_REG)/$(IMAGE_NAME)-proxy:$(IMAGE_TAG) --build-arg VERSION=$(VERSION)
	docker build . -f build/Dockerfile.controller -t $(IMAGE_REG)/$(IMAGE_NAME)-controller:$(IMAGE_TAG) --build-arg VERSION=$(VERSION)

push: check-vars ## 📤 Push container images
	@figlet $@ || true
	docker push $(IMAGE_REG)/$(IMAGE_NAME)-proxy:$(IMAGE_TAG)
	docker push $(IMAGE_REG)/$(IMAGE_NAME)-controller:$(IMAGE_TAG)

run-proxy: ## 🌐 Run proxy locally with hot-reload
	@figlet $@ || true
	@$(AIR_PATH) -c proxy/.air.toml

run-ctrl: ## 🤖 Run controller locally with hot-reload
	@figlet $@ || true
	@$(AIR_PATH) -c controller/.air.toml

test: ## 🧪 Run all unit tests
	@figlet $@ || true
	@echo "Not implemented yet! 😵"

clean: ## 🧹 Clean up, remove dev data and files
	@figlet $@ || true
	@rm -rf bin .tools tmp

release: ## 🚀 Release a new version on GitHub
	@figlet $@ || true
	@echo "💢 Releasing version $(VERSION) on GitHub"
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	gh release create "$(VERSION)" --title "v$(VERSION)" \
	--notes-file docs/release-notes.md \
	--latest 

helm-package: ## 🔠 Package Helm chart and update index
	@figlet $@ || true
	helm-docs --chart-search-root deploy/helm
	helm package deploy/helm/nanoproxy -d deploy/helm
	helm repo index deploy/helm
