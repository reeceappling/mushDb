#!/bin/bash
BACKUP_DIR="/backups/backup_$(date +%Y%m%d_%H%M%S)"

echo "Starting backup at $(date)"

mongodump --host="$MONGO_HOST" \
          --port="$MONGO_PORT" \
          --username="$MONGO_USER" \
          --password="$MONGO_PASS" \
          --authenticationDatabase=admin \
          --oplog \
          --gzip \
          --out="$BACKUP_DIR"

if [ $? -eq 0 ]; then
    echo "Backup completed successfully to $BACKUP_DIR"
else
    echo "Backup failed!"
fi

# Optional: Clean up backups older than 7 days
find /backups -type d -name "backup_*" -mtime +7 -exec rm -rf {} +