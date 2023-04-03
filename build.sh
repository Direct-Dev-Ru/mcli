#!/usr/bin/env bash

docker build -f .Dockerfiles/build/Dockerfile --target bin-linux --output bin-linux/ --platform linux/amd64 .
docker build -f .Dockerfiles/build/Dockerfile --target bin-windows --output bin-windows/ --platform windows/amd64 .