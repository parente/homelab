.PHONY: help clean cluster image venv

IMAGE:=ghcr.io/parente/cfsync
TAG?=latest

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## Clean up the workspace
	@rm -rf .venv

image: ## Build a local docker image
	@docker build --rm -t $(IMAGE):$(TAG) .

venv: ## Build a local dev virtualenv
	@python3 -m venv .venv
	@. .venv/bin/activate && pip install -r requirements.txt
