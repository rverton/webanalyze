#!/bin/bash
# https://raw.githubusercontent.com/tomnomnom/assetfinder/master/script/release

PROJDIR=$(cd `dirname $0`/.. && pwd)
DISTDIR=$PROJDIR/dist

VERSION="${1}"
TAG="v${VERSION}"
USER="rverton"
REPO="webanalyze"
REPO_PATH="cmd/webanalyze"
BINARY="${REPO}"

if [[ -z "${VERSION}" ]]; then
    echo "Usage: ${0} <version>"
    exit 1
fi

if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "You forgot to set your GITHUB_TOKEN"
    exit 2
fi

cd ${PROJDIR}

# Run the tests
go test
if [ $? -ne 0 ]; then
    echo "Tests failed. Aborting."
    exit 3
fi

# Check if tag exists
git fetch --tags
git tag | grep "^${TAG}$"

if [ $? -ne 0 ]; then
    github-release release \
        --user ${USER} \
        --repo ${REPO} \
        --tag ${TAG} \
        --name "${REPO} ${TAG}" \
        --description "${TAG}" \
        --pre-release
fi


for ARCH in "amd64" "386"; do
    for OS in "darwin" "linux" "windows" "freebsd"; do

        BINFILE="${BINARY}"

        if [[ "${OS}" == "windows" ]]; then
            BINFILE="${BINFILE}.exe"
        fi

        # rm -f ${BINFILE}

        GOOS=${OS} GOARCH=${ARCH} go build -o $DISTDIR/$BINFILE -ldflags "-X main.gronVersion=${VERSION}" github.com/${USER}/${REPO}/${REPO_PATH}

        if [[ "${OS}" == "windows" ]]; then
            ARCHIVE="${BINARY}-${OS}-${ARCH}-${VERSION}.zip"
            cd $DISTDIR && zip ${ARCHIVE} ${BINFILE}
        else
            ARCHIVE="${BINARY}-${OS}-${ARCH}-${VERSION}.tgz"
            cd $DISTDIR && tar --create --gzip -C $DISTDIR --file=${ARCHIVE} ${BINFILE}
        fi

        echo "Uploading ${ARCHIVE}..."
        github-release upload \
            --user ${USER} \
            --repo ${REPO} \
            --tag ${TAG} \
            --name "${ARCHIVE}" \
            --file ${DISTDIR}/${ARCHIVE}
    done
done
