.PHONY: all client server clean-client-build deploy-client

# Output binary name
BINARY_NAME=prism-server

# Default target: Build and deploy everything
all: deploy-client server

# Provision to build only the client (includes cleaning and building)
client: clean-client-build
	cd client && flutter build web --wasm --no-web-resources-cdn

# Provision to build only the server
server:
	cd server && go build -o ../$(BINARY_NAME) main.go

# Helper step: Remove the web folder in client build
clean-client-build:
	rm -rf client/build/web

# Helper step: Copy built client assets to server/web
# Depends on 'client' to ensure fresh build exists
deploy-client: client
	mkdir -p server/web
	rm -rf server/web/*
	cp -r client/build/web/* server/web/
