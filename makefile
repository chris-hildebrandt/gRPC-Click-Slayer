# -----------------------------
# Monster Slayer dev commands
# -----------------------------

# Go server port
SERVER_PORT=50051

# Envoy port
ENVOY_PORT=8080

# Envoy image version
ENVOY_IMAGE=envoyproxy/envoy:v1.29-latest

# Proto generation
proto:
	protoc \
	  --go_out=server/proto --go-grpc_out=server/proto \
	  --grpc-web_out=import_style=commonjs,mode=grpcwebtext:client/src/grpc \
	  -I=server/proto -I/usr/local/include \
	  server/proto/click.proto

# Start Go server
server:
	go run server/main.go

# Start Envoy (if not running)
envoy:
	@if [ $$(docker ps -q -f name=monsterslayer-envoy) ]; then \
		echo "Envoy already running"; \
	else \
		echo "Starting Envoy..."; \
		docker run -d --name monsterslayer-envoy -p $(ENVOY_PORT):8080 \
		-v $(PWD)/envoy.yaml:/etc/envoy/envoy.yaml $(ENVOY_IMAGE); \
	fi

# Stop Envoy
stop-envoy:
	docker stop monsterslayer-envoy && docker rm monsterslayer-envoy || true

# Run everything for development
dev: envoy server
