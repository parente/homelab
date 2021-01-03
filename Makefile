.PHONY: help apply clean cluster diff edit-secrets sync

SHARED_VOLUME?=${HOME}/docker/homelab
ARGS?=

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

apply: ## Apply helmfile changes to the cluster
	@helmfile -f helmfiles/helmfile.yaml --interactive apply $(ARGS)

cluster: ## Create a multi-host, multi-node cluster
# Pin server to 1.19.4 to avoid https://github.com/k3s-io/k3s/issues/2704
	@k3sup install \
		--ip 192.168.86.200 \
		--user pi \
		--ssh-key ~/.ssh/pacman \
		--k3s-channel v1.19 \
		--k3s-version 'v1.19.4+k3s2' \
		--k3s-extra-args '--no-deploy traefik --disable-cloud-controller' \
		--merge \
		--context pacman
	@k3sup join \
		--ip 192.168.86.201 \
		--server-ip 192.168.86.200 \
		--user pi \
		--ssh-key ~/.ssh/pacman \
		--k3s-channel v1.19
	@k3sup join \
		--ip 192.168.86.202 \
		--server-ip 192.168.86.200 \
		--user pi \
		--ssh-key ~/.ssh/pacman \
		--k3s-channel v1.19
	@k3sup join \
		--ip 192.168.86.203 \
		--server-ip 192.168.86.200 \
		--user pi \
		--ssh-key ~/.ssh/pacman \
		--k3s-channel v1.19 \
		--k3s-extra-args '--node-label homelab/external-device=zwave-controller'

diff: ## Diff the helmfile with the cluster
	@helmfile -f helmfiles/helmfile.yaml diff $(ARGS)

edit-secrets: ## Edit helm secrets in VSCode
	@EDITOR="code --wait" helm secrets edit helmfiles/secrets.yaml

localhost-cluster: ## Create a single-host, multi-node cluster
	@mkdir -p $(SHARED_VOLUME)
	@k3d cluster create homelab \
		-p "80:80@loadbalancer" \
		-p "443:443@loadbalancer" \
		--volume "$(SHARED_VOLUME):/var/lib/rancher/k3s/storage" \
		--k3s-server-arg '--no-deploy=traefik' \
		--agents 2

localhost-clean: ## Destroy a local cluster
	@k3d cluster delete homelab

sync: ## Sync the entire helmfile with the cluster
	@helmfile -f helmfiles/helmfile.yaml --interactive sync $(ARGS)
