#!/bin/bash

set -e

initialDir=$(pwd)
workingDir=$(git rev-parse --show-toplevel)
cd $workingDir

rm -rf bin/mushApi
GOOS=linux GOARCH=arm64 go build -o bin/mushApi .
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttp up --build --force-recreate
cd $initialDir