#!/usr/bin/env bash

docker buildx build --load -f ./.Dockerfiles/main-image/Dockerfile -t kuznetcovay/mcli:23.12.01 .