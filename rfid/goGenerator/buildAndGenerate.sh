#!/bin/bash

set -e
echo "building generator"
go build -o generateCode $(git rev-parse --show-toplevel)/rfid/goGenerator
echo "generating"
./generateCode
echo "formatting"
gofmt -w $(git rev-parse --show-toplevel)/rfid/generated.go
gofmt -w $(git rev-parse --show-toplevel)/rfid/generatedInterfaces.go
echo "cleaning up"
rm -f generateCode
