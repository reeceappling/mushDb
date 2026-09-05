#!/bin/bash

# Create a backup with default name
# sudo go run ./backups --create "/Users/ZNM4U7N/projects/other/mushDb/env/.env.prod"
# Create a backup with a specific name
# sudo go run ./backups --create --zipfile="/Users/ZNM4U7N/projects/other/mushDb/testBackups/backup.zip" "/Users/ZNM4U7N/projects/other/mushDb/env/.env.prod"

# Load a backup into test area
# sudo go run ./backups --load --zipfile="/Users/ZNM4U7N/projects/other/mushDb/testBackups/backup.zip" "/Users/ZNM4U7N/projects/other/mushDb/env/.env.load"