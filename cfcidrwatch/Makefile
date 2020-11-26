.PHONY: help builder get image imagex run

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

builder: ## Create a dockerx docker-container build for multi-arch build + push
	@docker buildx create --use --name build --node build --driver docker-container --driver-opt network=host

get: ## Install dependencies
	@go get -d -v

image: ## Build a local docker image
	@DOCKER_BUILDKIT=1 docker buildx build --load --tag cfcidrwatch:latest .

imagex: ## Make multi-arch platform images and push to GHCR as latest
	@DOCKER_BUILDKIT=1 docker buildx build \
		--push \
		--platform linux/amd64,linux/arm/v7 \
		--tag ghcr.io/parente/cfcidrwatch:latest .

run: ## Build and run locally
	@go run ./cfcidrwatch.go
