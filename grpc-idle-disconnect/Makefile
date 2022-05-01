.PHONY: all
all: go-server go-client swift-client

bin:
	mkdir bin

.PHONY: go-server
go-server: protos bin
	go build -o bin/go-server go-server/main.go

.PHONY: go-client
go-client: protos bin
	go build -o bin/go-client go-client/main.go

.PHONY: swift-client
swift-client: protos bin
	swift build
	cp `swift build --show-bin-path`/swift-client bin

service/service.pb.go: service/service.proto Makefile
	protoc	-I=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		service/service.proto

service/service.pb.swift service/service.grpc.swift: service/service.proto Makefile
	protoc	-I. \
		--swift_out=. \
		--swift-grpc_out=. \
		service/service.proto

.PHONY: protos
protos: service/service.pb.go service/service.pb.swift service/service.grpc.swift

.PHONY: clean
clean:
	go clean go-server/... go-client/...
	rm -rf bin .build
