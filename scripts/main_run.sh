#!/bin/bash

set -e

workingDir=$(git rev-parse --show-toplevel)
(
  cd $workingDir

  rm -rf bin/mushApi
  GOOS=linux GOARCH=arm64 go build -o bin/mushApi .
  # Only builds and launches 2
  #MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build web --build api --force-recreate
  MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --verbose --env-file env/.env.devhttps up --build --force-recreate
)