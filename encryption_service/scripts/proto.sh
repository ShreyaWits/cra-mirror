#/bin/bash

protoc   --proto_path=../internal/app/grpc/proto   --go_out=../internal/common/proto_gen --go_opt=paths=source_relative   --go-grpc_out=../internal/common/proto_gen --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional ../internal/app/grpc/proto/encryption.proto