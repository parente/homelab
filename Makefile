.PHONY: help clean cluster import sync

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## Destroy the cluster
	@k3d cluster delete k3dlab

cluster: ## Create a new cluster
	@k3d cluster create k3dlab \
		-p "80:80@loadbalancer" \
		-p "443:443@loadbalancer" \
		--k3s-server-arg '--no-deploy=traefik' \
		--agents 2

import: ## Build and import the local cfsync docker image for testing
	@make -C cfsync image
	@k3d image import ghcr.io/parente/cfsync:latest -c k3dlab

sync: ## Sync helmfile with the cluster
	@helmfile sync
