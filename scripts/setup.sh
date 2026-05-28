#!/bin/bash

initialDir=$(pwd)
workingDir=$(git rev-parse --show-toplevel)
cd $workingDir
mkdir -p ~/mush/pictures
chown -R $(id -u):$(id -g) ~/mush/pictures
chmod -R 777 ~/mush/pictures
cd $initialDir
