#!/bin/bash
set -e
ENV_NAME="$1"
case "$ENV_NAME" in
    "prod"|"Prod"|"PROD")
        # Commands to run if VARIABLE matches pattern1
        MNGO_DIR="/opt/mush/data/db"
        PICS_DIR="/opt/mush/data/pictures"
        ;;
    "cert"|"Cert"|"CERT")
        # Commands to run if VARIABLE matches pattern2 OR pattern3
        MNGO_DIR="/tmp/mush/data/cert/db"
        PICS_DIR="/tmp/mush/data/cert/pictures"
        ;;
    *) # |"dev"|"Dev"|"DEV"|"qual"|"Qual"|"QUAL"
        # Default fallback commands (like "default:" in C/Java)
        MNGO_DIR="/tmp/mush/data/db"
        PICS_DIR="/tmp/mush/data/pictures"
        ;;
esac

prep_dir() {
    local filepath="$1"
    echo "setting up path: $filepath"
    # Create the dir
    mkdir -p $filepath
    # Set the ownership to the current user and group
    chown -R $(id -u):$(id -g) $filepath
    # Allow all to read and write
    chmod -R a+rw $filepath
}

prep_dir $MNGO_DIR
prep_dir $PICS_DIR

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
#chmod -R a+rw picDir


