#!/bin/bash

# Variables
VERSION_FILE="VERSION.txt"

GITHUB_TOKEN="${GITHUB_TOKEN}"  # Replace with your GitHub token

REPO="direct-dev-ru/binaries"  # Replace with your GitHub username/repo

TAG=swknf.$(cat "$VERSION_FILE")

echo TAG: $TAG

RELEASE_DIR="/home/su/projects/golang/cobra-cli-example/for-upload"  

body="{\"tag_name\":\"${TAG}\", \"target_commitish\":\"main\", \"name\":\"${TAG}\", \
  \"body\":\"${TAG}\", \"draft\":false, \"prerelease\":false, \"generate_release_notes\":false}"

echo BODY: $body

response=$(curl -L -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/direct-dev-ru/binaries/releases \
  -d $body)

echo $response

# Extract the upload URL from the response
upload_url=$(echo "$response" | jq -r '.upload_url' | sed "s/{?name,label}//")

# Check if the release was created successfully
if [[ "$response" == *"Not Found"* ]]; then
    echo "Error: Repository not found or invalid token."
    exit 1
fi

# Upload each binary file
for file in "$RELEASE_DIR"/*; do
    if [[ -f "$file" ]]; then
        filename=$(basename "$file")
        echo "Uploading $filename..."
        response=$(curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
            -H "Content-Type: application/octet-stream" \
            "$upload_url?name=$filename" \
            --data-binary @"$file")
        echo $response    
    fi
done

echo "All binaries uploaded successfully."
