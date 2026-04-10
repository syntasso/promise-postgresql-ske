IMAGE ?= ghcr.io/syntasso/promise-postgresql-ske/postgresql-configure-pipeline
TAG   ?= v1.0.0 # x-release-please-version
KIND_CLUSTERS ?= platform,worker

BUILD_FLAGS := -t $(IMAGE):$(TAG)
ifdef CI
BUILD_FLAGS += --cache-from type=gha --cache-to type=gha,mode=max
endif

ROOT   ?= $(abspath ../..)
WORKER_CONTEXT ?= kind-worker

PIPELINE_TESTS_DIR := test/resources

.PHONY: help build clean load pull push test system-test

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

build: ## Build the Docker image
	docker buildx build $(BUILD_FLAGS) --load .

clean: ## Remove the Docker image from the local machine
	docker rmi $(IMAGE):$(TAG)

load: build ## Load the Docker image into the kind cluster
	for cluster in $(shell echo $(KIND_CLUSTERS) | tr ',' ' '); do \
		kind load docker-image $(IMAGE):$(TAG) --name $$cluster; \
	done

pull: ## Pull the Docker image from the registry
	docker pull $(IMAGE):$(TAG)

push: ## Build and push the Docker image for linux/amd64 and linux/arm64
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE):$(TAG) --push .

test: build ## Run kratix pipeline tests for the resource configure workflow
	kratix test pipeline resource/configure/instance-configure \
		--dir $(ROOT) \
		--input-dir $(PIPELINE_TESTS_DIR)/inputs \
		--expected-output-dir $(PIPELINE_TESTS_DIR)/expected-outputs

system-test: ## Run Ginkgo system tests
	WORKER_CONTEXT=$(WORKER_CONTEXT) \
	PROMISE_YAML=$(ROOT)/promise.yaml \
	PROMISE_HA_YAML=$(ROOT)/promise-ha.yaml \
	RESOURCE_REQUEST_YAML=$(ROOT)/resource-request.yaml \
	MULTIPLE_RESOURCE_REQUESTS_YAML=$(ROOT)/multiple-resource-requests.yaml \
	go test -v ./test/... -timeout 10m
