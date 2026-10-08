# кодогенерации на основе openAPI контракта по всем контрактам в папке api
gen:
	@for s in auth dict candidate employer testing interaction search; do \
		echo ">> $$s"; \
		go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
			-config api/$$s/codegen.yaml api/$$s/openapi.yaml; \
	done