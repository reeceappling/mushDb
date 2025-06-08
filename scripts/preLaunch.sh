#!/bin/bash

# Setup cleanup function
ORIGINAL_DIR=$(pwd)
function cleanup {
  cd $ORIGINAL_DIR
}
trap cleanup EXIT

# Move into the repo's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$SCRIPT_DIR"
cd ../
REPO_DIR=$(pwd)

cp initDB/mongo-init.js /tmp/mush/mongo-init.js
echo "moved db init script to /tmp/mush/mongo-init.js"