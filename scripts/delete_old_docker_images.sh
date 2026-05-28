#!/bin/bash

set -e

# Remove containers without names and tags
docker image prune -f

# Remove all volumes
# docker volume prune -a

# Remove docker builds that are not tagged and not used by any container
docker builder prune -f
# Remove all docker builds
# docker builder prune -a -f

#docker images --format json | grep "mush-web"
#docker images --format json | grep "mush-api"
#docker images --format json | jq '.[] | select(.Repository == "mush-web" or .status == "mush-api")'