fetch-openapi:
	curl -fsSL https://api-docs.humanitec.com/openapi.json > ./client/openapi.json

generate:
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,client -package client ../api-docs/output/humanitec.json > client/client.gen.go
