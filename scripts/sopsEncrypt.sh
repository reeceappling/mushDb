#!/bin/bash
# Requires files to exist: env/.env.prod, etc/mongodb.key
# Req: 1 arg that is the sops public keystring. Run from project root dir.
SOPS_PUB_KEY=$1
sops --encrypt --input-type dotenv --output-type dotenv --age $SOPS_PUB_KEY env/.env.prod > env/prod-enc.env
sops --encrypt --age $SOPS_PUB_KEY  etc/mongodb.key > etc/mongodb-enc.key