#!/bin/bash
# Run from
set -e
echo "building generator"
go build -o generateCode $(git rev-parse --show-toplevel)/api/goGenerator
echo "generating"
./generateCode
echo "formatting"
gofmt -w $(git rev-parse --show-toplevel)/api/generated.go
gofmt -w $(git rev-parse --show-toplevel)/api/generatedInterfaces.go
echo "cleaning up"
rm -f generateCode
