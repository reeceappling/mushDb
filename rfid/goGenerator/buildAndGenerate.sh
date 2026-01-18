#!/bin/bash

set -e
echo "building generator"
go build -o generateCode $(git rev-parse --show-toplevel)/rfid/goGenerator
echo "generating"
./generateCode
echo "cleaning up"
rm -f generateCode
