#!/bin/bash

set -e

workingDir=$(git rev-parse --show-toplevel)
(
  cd $workingDir
  rm -rf bin/mushApi
  GOOS=linux GOARCH=arm64 go build -o bin/mushApi .
  MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --verbose --env-file env/.env.prod up --build --force-recreate
)