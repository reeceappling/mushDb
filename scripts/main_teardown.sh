#!/bin/bash

initialDir=$(pwd)
workingDir=$(git rev-parse --show-toplevel)

cd $workingDir
docker compose --env-file env/.env.devhttp down
rm -rf bin/mushApi
cd $initialDir