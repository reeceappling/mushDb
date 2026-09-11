#!/bin/bash
PASSWORD=$1
INPUT_FILE=$2
OUTPUT_FILE=$3

if [ -z "$PASSWORD" ] || [ -z "$INPUT_FILE" ] || [ -z "$OUTPUT_FILE" ]; then
  echo "Usage: $0 <password> <input_file> <output_file>"
  exit 1
fi
openssl enc -aes-256-cbc -pbkdf2 -d -in "$INPUT_FILE" -out "$OUTPUT_FILE" -pass "pass:$PASSWORD"
echo "File decrypted successfully to $OUTPUT_FILE"