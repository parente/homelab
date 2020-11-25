.PHONY: help clean cluster image venv

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## Clean up the workspace
	@rm -rf .venv

image: ## Build a local docker image
	@DOCKER_BUILDKIT=1 docker buildx build --load --tag cfsync:latest .

imagex: ## Make multi-arch platform images and push to GHCR as latest
	@DOCKER_BUILDKIT=1 docker buildx build \
		--push \
		--platform linux/amd64,linux/arm/v7 \
		--tag ghcr.io/parente/cfsync:latest .

venv: ## Build a local dev virtualenv
	@python3 -m venv .venv
	@. .venv/bin/activate && pip install -r requirements.txt
