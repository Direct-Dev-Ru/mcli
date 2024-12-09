#!/bin/bash

docker build -f Dockerfiles/LocalCompile/Dockerfile --target bin-linux --output bin-linux-amd64/ --platform linux/amd64 .
docker build -f Dockerfiles/LocalCompile/Dockerfile --target bin-linux --output bin-linux-arm64/ --platform linux/arm64 .

# in linux setuid
# sudo chown root:root bin-linux/lcg
# sudo chmod +s bin-linux/lcg