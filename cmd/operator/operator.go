//go:build tools
// +build tools

package main

// protobuf generators
import (
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
