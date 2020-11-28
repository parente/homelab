.PHONY: help apply clean cluster diff sync

SHARED_VOLUME?=${HOME}/docker/homelab

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

apply: ## Apply helmfile changes to the cluster
	@helmfile apply

clean: ## Destroy the cluster
	@k3d cluster delete homelab

cluster: ## Create a new cluster
	@mkdir -p $(SHARED_VOLUME)
	@k3d cluster create homelab \
		-p "80:80@loadbalancer" \
		-p "443:443@loadbalancer" \
		--volume "$(SHARED_VOLUME):/var/lib/rancher/k3s/storage" \
		--k3s-server-arg '--no-deploy=traefik' \
		--agents 2

diff: ## Diff the helmfile with the cluster
	@helmfile diff

edit-secrets: ## Edit helm secrets in VSCode
	@EDITOR="code --wait" helm secrets edit secrets.yaml

sync: ## Sync the entire helmfile with the cluster
	@helmfile sync
