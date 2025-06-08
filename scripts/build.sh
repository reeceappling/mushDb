#!/bin/bash

# TODO: FIX DOCKERFILES
# TODO: FIX HELM CHARTS

# Default variable values
# verbose_mode=false # TODO: ??????
# output_file="" # TODO: ??????

# Function to display script usage
usage() {
 echo "Usage: $0 [OPTIONS] [TO_BUILD]"
 echo "Options:"
 echo " -h, --help      Display this help message"
 echo " -v, --verbose   Enable verbose mode"
 echo "TO_BUILD:"
 echo " api      Build the main server (go api)"
 echo " web      Build the webserver (react api)"
 echo " rfid     Build the rfid-reader/writer client"
}

has_argument() {
    [[ ("$1" == *=* && -n ${1#*=}) || ( ! -z "$2" && "$2" != -*)  ]];
}

extract_argument() {
  echo "${2:-${1#*=}}"
}

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      -h | --help)
        usage
        exit 0
        ;;
      -v | --verbose)
        verbose_mode=true
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

# Setup cleanup function
ORIGINAL_DIR=$(pwd)
function cleanup {
  cd $ORIGINAL_DIR
}
trap cleanup EXIT

# TODO: use next line or delete
makeBinDir() {
    mkdir bin
}

# Move into the repo's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$SCRIPT_DIR"
cd ../
REPO_DIR=$(pwd)

case $1 in
  "api")
    rm -f bin/mushApi
    echo "compiling api"
    GOOS=linux GOARCH=arm64 go build -o bin/mushApi .
    chmod 0777 bin/mushApi
    # Remove old image
    echo "removing old image"
    docker rmi mush-api:latest
    echo "building api image mush-api:latest"
    docker build -t mush-api:latest -f ./DockerfileApi .
    rm -f bin/mushApi
    ;;

  "web")
    cd rfid/client
    # Remove old image
    echo "removing old image"
    docker rmi mush-web:latest
    echo "setting build args"
    export MAIN_API_EXTERNAL_HOST=home.appli.ng
    export MAIN_API_INTERNAL_HOST=localhost:80
    echo "building webserver image mush-web:latest"
    docker build -t mush-web:latest -f ./Dockerfile --build-arg MAIN_API_EXTERNAL_HOST --build-arg MAIN_API_INTERNAL_HOST .
    ;;

  "rfid")
    # TODO: DOWNLOAD REPO?
    # TODO: BUILD FROM REPO?
    # TODO: remove old docker image?
    # TODO: build/tag docker image?
    ;;

  *)
    echo "invalid build target"
    exit 1
    ;;
esac

#!/bin/bash



# Now you are in the script's directory and can execute commands relative to it
# ... your script commands here ...