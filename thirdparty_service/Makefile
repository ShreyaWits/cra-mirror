run: build protos
	@./bin/main

build: protos
	@go generate ./...
	@go build -o bin/main main.go

protos:
	@protoc \
      --proto_path=./proto \
      --go_out=./proto \
      --go-grpc_out=./proto \
      --go_opt=paths=source_relative \
      --go-grpc_opt=paths=source_relative \
      ./proto/thirdparty.proto

.PHONY: build run protos
