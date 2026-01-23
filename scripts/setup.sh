#!/bin/bash

initialDir=$(pwd)
workingDir=$(git rev-parse --show-toplevel)
cd $workingDir
mkdir -p ~/mush/pictures
cp testImages/test.jpg ~/mush/pictures/test.jpg
chown -R $(id -u):$(id -g) ~/mush/pictures
chmod -R 777 ~/mush/pictures
cd $initialDir
