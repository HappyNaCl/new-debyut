PROTO_DIR   := api/protobuf
OUT_DIR     := internal/genproto
MODULE      := github.com/HappyNaCl/new-debyut

PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

GATEWAY_MAIN := cmd/api-gateway/main.go
SWAGGER_DIR  := api/swagger

.PHONY: proto proto-clean swagger

proto:
	protoc --go_out=. --go_opt=module=$(MODULE) --go-grpc_out=. --go-grpc_opt=module=$(MODULE) $(PROTO_FILES)

swagger:
	swag init -g $(GATEWAY_MAIN) -o $(SWAGGER_DIR) --parseInternal

proto-clean:
	del /s /q $(subst /,\,$(OUT_DIR))\*.go 2>nul

migrate:
	go run .\cmd\run-migrations

seed: 
	go run .\cmd\run-migrations --seed