run: build protos
	@./bin/main

build: protos
	@go generate ./...
	@go build -o bin/main main.go

protos:
	@protoc \
      --proto_path=protos \
      --go_out=internal/pb \
      --go-grpc_out=internal/pb \
      --go_opt=paths=source_relative \
      --go-grpc_opt=paths=source_relative \
      hello.proto

.PHONY: build run protos
