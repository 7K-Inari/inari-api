OAPI_CODEGEN ?= oapi-codegen

.PHONY: generate lint breaking build test ci drift

generate:
	buf generate
	$(OAPI_CODEGEN) -config openapi/oapi-codegen.yaml openapi/openapi.yaml
	npm run generate:rest

lint:
	buf lint
	npx redocly lint openapi/openapi.yaml

breaking:
	buf breaking --against ".git#branch=origin/main"

build:
	go build ./...
	go vet ./...
	npx tsc --noEmit

test:
	go test ./...

drift: generate
	git diff --exit-code

ci: lint build test
