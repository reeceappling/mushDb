#!/bin/bash
set -e
ENV_NAME="$1"
case "$ENV_NAME" in
    "prod"|"Prod"|"PROD")
        rm -rf /opt/mush/data/db/*
        rm -rf /opt/mush/data/pictures/*
        ;;
    "cert"|"Cert"|"CERT")
        # Commands to run if VARIABLE matches pattern2 OR pattern3
        rm -rf /tmp/mush/data/cert/db/*
        rm -rf /tmp/mush/data/cert/pictures/*
        ;;
    *) # |"dev"|"Dev"|"DEV"|"qual"|"Qual"|"QUAL"
        # Default fallback commands (like "default:" in C/Java)
        rm -rf /tmp/mush/data/db/*
        rm -rf /tmp/mush/data/pictures/*
        ;;
esac