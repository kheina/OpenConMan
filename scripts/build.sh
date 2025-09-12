#!/usr/bin/env bash

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+changes" || true)"
GIT_BRANCH="$(git branch --show-current)"
VERSION="$(<VERSION)"
TIME="$(date +%s)"
GOPATH=${GOPATH:-$(go env GOPATH)}

if [ "${OPENCONMAN_BUILD_UI}x" != "x" ]; then
	echo "==> Building UI assets..."
	cd website && npm run build
	cd ..
fi

if [ "${GOOS}x" == "x" ]; then
	GOOS=$(go env GOOS)
fi
if [ "${GOARCH}x" == "x" ]; then
	GOARCH=$(go env GOARCH)
fi

echo "==> Building OpenConMan for ${GOOS}/${GOARCH}..."

go build \
-o bin/conman \
-ldflags "
	-X 'github.com/kheina/openconman/src/version.Commit=${GIT_COMMIT}${GIT_DIRTY}'
	-X 'github.com/kheina/openconman/src/version.Branch=${GIT_BRANCH}'
	-X 'github.com/kheina/openconman/src/version.VersionStr=${VERSION}'
	-X 'github.com/kheina/openconman/src/version.Timestamp=${TIME}'
" \
./src

if [ "${OPENCONMAN_INSTALL_BIN}x" != "x" ]; then
	echo "==> Moving binary into GOPATH/bin..."
	mv -f bin/conman "${GOPATH}/bin/"
else
	echo "==> Built binary as bin/conman"
fi

echo "done."
