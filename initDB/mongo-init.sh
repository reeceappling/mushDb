#!/bin/sh
mongo <<EOF
use admin
db.createUser({
  user: '$MONGO_INITDB_ROOT_USERNAME',
  pwd: '$MONGO_INITDB_ROOT_PASSWORD',
  roles: [
    { role: 'readAnyDatabase', db: "admin" },
    { role: 'dbAdminAnyDatabase', db: "admin" },
    { role: 'userAdminAnyDatabase', db: "admin" },
    { role: 'readWrite', db: '$MONGO_INITDB_DATABASE' }
  ]
})
db.createUser({
  user: '$MONGO_INITDB_USERNAME',
  pwd: '$MONGO_INITDB_PASSWORD',
  roles: [
    { role: 'readAnyDatabase', db: "admin" },
    { role: 'dbAdminAnyDatabase', db: "admin" },
    { role: 'userAdminAnyDatabase', db: "admin" },
    { role: 'dbOwner', db: '$MONGO_INITDB_DATABASE' },
    { role: 'readWrite', db: '$MONGO_INITDB_DATABASE' }
  ]
})
db.createCollection("users")
db.createCollection("fruits")
db.createCollection("transfers")
EOF