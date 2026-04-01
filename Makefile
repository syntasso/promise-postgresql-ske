WORKFLOW_DIR   := workflows/configure-pipeline
WORKER_CONTEXT ?= kind-worker

.PHONY: help build-workflow load-workflow push-workflow test

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'

build-workflow: ## Build workflow pipeline image
	$(MAKE) -C $(WORKFLOW_DIR) build

load-workflow: ## Build and load the workflow pipeline image into kind
	$(MAKE) -C $(WORKFLOW_DIR) load

push-workflow: ## Build and push the workflow pipeline image to the registry
	$(MAKE) -C $(WORKFLOW_DIR) push

test: ## Run all tests
	$(MAKE) -C $(WORKFLOW_DIR) test WORKER_CONTEXT=$(WORKER_CONTEXT)
