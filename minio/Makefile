.PHONY: help buckets users www

help:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z0-9_%/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

buckets: ## Create public minio static site buckets
	@mc mb -p homelab/www
	@mc policy set download homelab/www
	@mc mb -p homelab/home
	@mc policy set download homelab/home
	@mc mb -p homelab/home/juliana
	@mc mb -p homelab/home/peter

users: ## Create minio users with personal buckets
	@mc admin user add juliana
	@mc admin user add peter
	@mc admin group add homelab family juliana
	@mc admin policy add homelab homeowner ./minio/homeowner-policy.json
	@mc admin policy set homelab homeowner group=family

www: ## Upload www assets
	@mc cp --recursive www/ homelab/www/