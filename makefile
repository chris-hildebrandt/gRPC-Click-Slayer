# -----------------------------
# gRPC Monster Slayer dev commands
# -----------------------------

# Go server port
SERVER_PORT=50051

# Envoy port
ENVOY_PORT=8081

# Proto generation
proto:
	protoc \
	  --go_out=server/proto --go-grpc_out=server/proto \
	  --grpc-web_out=import_style=commonjs,mode=grpcwebtext:client/src/grpc \
	  -I=server/proto -I/usr/local/include \
	  server/proto/click.proto

# Start Go gRPC server
server:
	cd server && go run main.go

# Start React client
client:
	cd client && npm install && npm run dev

# Install dependencies
deps:
	cd server && go mod tidy
	cd client && npm install

# Clean up
clean:
	rm -f server/players.json
	rm -rf client/node_modules
	rm -rf client/dist
	rm -rf client/src/grpc/*.js
