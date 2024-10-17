fetch-openapi:
	curl -fsSL https://api-docs.humanitec.com/openapi.json > ./docs/openapi.json

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 -generate types,client -package client ./docs/openapi.json > client/client.gen.go

test:
	go test ./... -cover
