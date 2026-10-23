#!/bin/bash
workingDir=$(git rev-parse --show-toplevel)
(
cd $workingDir
docker compose --env-file env/.env.devhttps down
docker compose --env-file env/.env.cert down
docker compose --env-file env/.env.prod down
rm -rf bin/mushApi
)