#!/usr/bin/env bash

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

UNIX_TIME="$(date +%s)"
GIT_BRANCH="$(git branch --show-current)"
VERSION="$(<VERSION)"

if [ "${GIT_BRANCH}x" != "mainx" ]; then
	VERSION=${VERSION}+dev-${UNIX_TIME}
fi

echo ${VERSION}
