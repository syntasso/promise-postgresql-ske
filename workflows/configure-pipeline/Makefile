IMAGE ?= ghcr.io/syntasso/ske-promise-example-postgresql/postgresql-configure-pipeline
TAG   ?= v0.0.0 # x-release-please-version
KIND_CLUSTER ?= platform

BUILD_FLAGS := -t $(IMAGE):$(TAG)
ifdef CI
BUILD_FLAGS += --cache-from type=gha --cache-to type=gha,mode=max
endif

ROOT   ?= $(abspath ../..)
WORKER_CONTEXT ?= kind-worker

.PHONY: help build clean load pull push test

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

build: ## Build the Docker image
	docker buildx build $(BUILD_FLAGS) --load .

clean: ## Remove the Docker image from the local machine
	docker rmi $(IMAGE):$(TAG)

load: build ## Load the Docker image into the kind cluster
	kind load docker-image $(IMAGE):$(TAG) --name $(KIND_CLUSTER)

pull: ## Pull the Docker image from the registry
	docker pull $(IMAGE):$(TAG)

push: ## Build and push the Docker image for linux/amd64 and linux/arm64
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE):$(TAG) --push .

test: ## Run Ginkgo system tests
	WORKER_CONTEXT=$(WORKER_CONTEXT) \
	PROMISE_YAML=$(ROOT)/promise.yaml \
	RESOURCE_REQUEST_YAML=$(ROOT)/resource-request.yaml \
	go test -v ./test/... -timeout 10m
