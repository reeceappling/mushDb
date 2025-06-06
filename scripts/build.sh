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
  rm  -r /tmp/foo
}
trap cleanup EXIT

makeBinDir() {
    mkdir bin
}

# Move into the repo's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$SCRIPT_DIR"
cd ../
REPO_DIR=$(pwd)


TO_BUILD=
case $1 in
  "api")
    rm -f bin/mushApi
    go build -o bin/mushApi ./rfid
    chmod 0777 bin/mushApi
    # TODO: remove old docker image?
    docker build -t mushDb:latest -f DockerfileApi
    ;;

  "web")
    cd rfid/client
    docker build -t mushWeb:latest -f Dockerfile
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