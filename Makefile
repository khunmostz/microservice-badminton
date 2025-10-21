PROTO_FILES := $(shell find proto -name "*.proto")

.PHONY: proto up down

proto:
	protoc -I proto --go_out=./proto --go_opt=paths=source_relative \
		--go-grpc_out=./proto --go-grpc_opt=paths=source_relative $(PROTO_FILES)

up:
	docker compose up -d --build

down:
	docker compose down -v
