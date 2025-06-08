#!/bin/bash

set -e

partName=$1
imgName=mush-$partName

sh scripts/build.sh $partName

docker save $imgName > $imgName.tar
multipass transfer $imgName.tar microk8s-vm:/tmp/$imgName.tar
microk8s ctr image import /tmp/$imgName.tar
rm -f $imgName.tar