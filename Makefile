OAPI_CODEGEN ?= oapi-codegen

.PHONY: generate lint breaking build test ci drift import-openapi

# Pull the full REST surface from inari-server's offline export and merge it
# with openapi/meta.fragment.yaml into openapi/openapi.yaml. Optionally pass a
# pre-exported spec: make import-openapi EXPORTED=/path/to/openapi.yaml
import-openapi:
	scripts/import-openapi.sh $(EXPORTED)

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
