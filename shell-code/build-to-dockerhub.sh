#!/bin/bash

REPO=kuznetcovay/go-lcg
VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION=v1.0.8    
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

# Run go tests
if ! go test -v -run=^Test; then
        echo "Tests failed. Exiting..."
        exit 1
fi

# Push multi-platform images
docker buildx build --push --platform linux/amd64,linux/arm64 -t ${REPO}:"${VERSION}" . ||
        {
                echo "docker buildx build --push failed. Exiting with code 1."
                exit 1
        }
