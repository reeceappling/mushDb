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

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      "api")
        sh scripts/internal/buildAndMove.sh api
        ;;
      "web")
        sh scripts/internal/buildAndMove.sh web
        ;;
      *)
        echo "Invalid option: $1" >&2
        usage
        exit 1
        ;;
    esac
    # Move arguments
    shift
  done
}

helm uninstall test-chart
# Main script execution
handle_options "$@"

sleep 10
helm install -f noCommit-testvalues.yaml test-chart fungi-tracker
sleep 5
podName=$(kubectl get pods | grep mushrooms-app- | cut -d" " -f1)
sleep 5
kubectl describe pods

#kubectl logs $(kubectl get pods | grep mushrooms-app- | cut -d" " -f1) mush-db
kubectl get pvc
#kubectl describe pvc
sleep 5
kubectl logs $podName api
