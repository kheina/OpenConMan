#!/usr/bin/env bash

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

echo "$(go version)"

echo "==> Installing Go Dependencies..."
go mod download

echo "==> Verifying Go Dependencies..."
go mod verify

echo "using node version $(node -v)"
echo "==> Installing Node Dependencies..."
cd website && npm install
cd ..

echo "done."
