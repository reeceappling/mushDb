#!/bin/bash

initialDir=$(pwd)
workingDir=$(git rev-parse --show-toplevel)
cd $workingDir
mkdir -p /tmp/mush/data/db
mkdir /tmp/mush/data/pictures
chown -R $(id -u):$(id -g) /tmp/mush/data/db
chown -R $(id -u):$(id -g) /tmp/mush/data/pictures
# Whitelist the Parent Directory in Docker Desktopmac
# OS requires explicit permission for Docker to access local files.
# 1. Open Docker Desktop.Click the Settings (gear icon) in the top right.
# 2. Go to Resources > File Sharing.
# 3. Click the + (plus) button and add the parent directory (or your ~ home directory).
# 4. Click Apply & restart to save the changes.

# TO FIGURE OUT A STOPPED CONTAINER's INTERNAL UID:
# docker compose run --rm <service_name> id -u
# If blank: The container is defaulting to root (UID 0).
# If a number: This is the exact UID the container forces at runtime.
# If a username: The container maps to a name defined in its internal /etc/passwd file.

# Run this command on your Mac to give all container users read/write access (useful if you are unsure of the container's internal UID
chmod -R a+rw /tmp/mush/data/db
chmod -R a+rw /tmp/mush/data/pictures
#mkdir -p ~/mush/pictures
#chown -R $(id -u):$(id -g) ~/mush/pictures
#chmod -R 777 ~/mush/pictures
#cd $initialDir
