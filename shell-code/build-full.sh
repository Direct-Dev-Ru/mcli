#!/bin/bash

# REPO=registry.direct-dev.ru/go-lcg
REPO=kuznetcovay/swknf
VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION=v1.1.1
fi
BRANCH=main

echo "${VERSION}" > VERSION.txt

export GOCACHE="${HOME}/.cache/go-build"

# Save the current branch
CURRENT_BRANCH=$(git branch --show-current)

# Function to restore the original branch
function restore_branch {
        echo "Restoring original branch: ${CURRENT_BRANCH}"
        git checkout "${CURRENT_BRANCH}"
}

# Check if the current branch is different from the target branch
if [ "$CURRENT_BRANCH" != "$BRANCH" ]; then
        # Set a trap to restore the branch on exit
        trap restore_branch EXIT
        echo "Switching to branch: ${BRANCH}"
        git checkout ${BRANCH}
fi

# Fetch all tags from the remote repository
git fetch --tags

# Check if the specified version tag exists
if git rev-parse "refs/tags/${VERSION}" >/dev/null 2>&1; then
        echo "Tag ${VERSION} already exists. Halting script."
        exit 1
fi

# Run go tests
# if ! go test -v -run=^Test; then
#         echo "Tests failed. Exiting..."
#         exit 1
# fi
mkdir -p for-upload

# Build for linux/amd64
docker build -f Dockerfiles/LocalCompile/Dockerfile --target bin-linux --output bin-linux-amd64/ --platform linux/amd64 . ||
        {
                echo "docker build for amd64 failed. Exiting with code 1."
                exit 1
        }

cp bin-linux-amd64/swknf  "upload/swknf.amd64.${VERSION}"

# Build for linux/arm64
docker build -f Dockerfiles/LocalCompile/Dockerfile --target bin-linux --output bin-linux-arm64/ --platform linux/arm64 . ||
        {
                echo "docker build for arm64 failed. Exiting with code 1."
                exit 1
        }

cp bin-linux-arm64/swknf  "upload/swknf.arm64.${VERSION}"

# Push multi-platform images
docker buildx build -f Dockerfiles/ImageBuild/Dockerfile --push --platform linux/amd64,linux/arm64 -t ${REPO}:"${VERSION}" . ||
        {
                echo "docker buildx build --push failed. Exiting with code 1."
                exit 1
        }

git add -A . ||
        {
                echo "git add failed. Exiting with code 1."
                exit 1
        }

git commit -m "release $VERSION" ||
        {
                echo "git commit failed. Exiting with code 1."
                exit 1
        }

git tag -a "$VERSION" -m "release $VERSION" ||
        {
                echo "git tag failed. Exiting with code 1."
                exit 1
        }
        
git push -u origin main --tags ||
        {
                echo "git push failed. Exiting with code 1."
                exit 1
        }

