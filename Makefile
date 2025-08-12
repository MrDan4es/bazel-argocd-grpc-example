.PHONY: package

package: ## Сборка umbrella helm oci.
	helm dependency build helm/umbrella/ && helm package helm/umbrella

help: ## Показать помощь
	@echo "Доступные цели:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m- %-8s\033[0m \033[32m%s\033[0m\n", $$1, $$2}'

.DEFAULT_GOAL := help
