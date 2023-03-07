fetch-openapi:
	curl -fsSL https://dev-update-docs-api-docs.humanitec.io/openapi.json > ./docs/openapi.json

generate:
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,client -package client ./docs/openapi.json > client/client.gen.go

test:
	go test ./... -cover
