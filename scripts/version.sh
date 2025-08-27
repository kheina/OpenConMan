SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

GIT_BRANCH="$(git branch --show-current)"
VERSION="$(<VERSION)"

if [ "${GIT_BRANCH}x" != "mainx" ]; then
	VERSION=${VERSION}+dev
fi

echo ${VERSION}
