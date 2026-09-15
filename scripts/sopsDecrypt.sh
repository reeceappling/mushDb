#!/bin/bash
# Req: 1 arg that is the sops private key as a string.  Run from project root dir.
SOPS_AGE_KEY=$1
sops --decrypt --input-type dotenv --output-type dotenv env/prod-enc.env > env/.env.prod
sops --decrypt etc/mongodb-enc.key > etc/mongodb.key