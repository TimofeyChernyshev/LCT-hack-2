# кодогенерации на основе openAPI контракта по всем контрактам в папке api
CONTRACTS := auth dict candidate employer testing interaction search
FE_API_DIR := front/src/api

gen: gen-go gen-frontend

gen-go: ## Go: oapi-codegen по всем api/<service>/openapi.yaml
	@go get github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	
	@for s in $(CONTRACTS); do \
		echo ">> $$s"; \
		go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
			-config api/$$s/codegen.yaml api/$$s/openapi.yaml; \
	done

	@go mod tidy

gen-frontend: ## TS: openapi-generator-cli (через docker, без Java в хосте)
	@mkdir -p $(FE_API_DIR)
	@for s in $(CONTRACTS); do \
		echo ">> ts  $$s"; \
		docker run --rm \
			-v "$$PWD:/local" \
			openapitools/openapi-generator-cli:v7.10.0 generate \
				-i /local/api/$$s/openapi.yaml \
				-g typescript-fetch \
				-o /local/$(FE_API_DIR)/$$s \
				--additional-properties=\
supportsES6=true,\
npmName=@fsp/api-$$s,\
npmVersion=0.0.1,\
withInterfaces=true,\
useSingleRequestParameter=true; \
	done