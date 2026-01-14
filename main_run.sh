#!/bin/bash

set -e

rm -rf bin/mushApi
# -x on build?
GOOS=linux GOARCH=arm64 go build -o bin/mushApi .
docker compose --env-file env/.env.devhttp up --build --force-recreate