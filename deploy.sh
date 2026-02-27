!/usr/bin/env bash

set -e

FRONTEND_DIR="/home/lee/open_disc/open_disc/frontend"
NGINX_ROOT="/var/www/html"
BACKEND_DIR="/home/lee/open_disc/open_disc"

echo "Moving to frontend root"
cd "$FRONTEND_DIR"

echo "Installing dependencies"
bun install

echo "building frontend"
bun run build

echo "Syncing build output to nginx root"
sudo rsync -av --delete "$FRONTEND_DIR/dist"/ "$NGINX_ROOT"/

echo "Done deploying frontend"

echo "Moving to backend root"
cd "$BACKEND_DIR"

echo "Building backend executable"
go build internal/main/main.go

echo "Running backend"
./main
