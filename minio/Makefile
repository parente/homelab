.PHONY: help buckets users www

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

buckets: ## Create minio buckets
	@mc mb -p homelab/www
	@mc policy set download homelab/www
	@mc mb -p homelab/home
	@mc policy set download homelab/home
	@mc mb -p homelab/home/juliana
	@mc mb -p homelab/home/peter
	@mc mb -p homelab/backup/webthings
	@mc version enable homelab/backup/webthings

users: ## Create minio users with personal buckets
	@mc admin user add homelab juliana
	@mc admin group add homelab family juliana
	@mc admin policy add homelab homeowner ./homeowner-policy.json
	@mc admin policy set homelab homeowner group=family

backup: ## Create minio backup policy and user
	@mc admin user add homelab backup
	@mc admin policy add homelab backup ./backup-policy.json
	@mc admin policy set homelab backup user=backup

www: ## Upload www assets
	@mc cp --recursive www/ homelab/www/