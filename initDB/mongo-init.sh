#!/bin/sh
mongo <<EOF
use admin
// The first user is for the root user, only used by the database itself. ROOT USER ALREADY EXISTS
//db.createUser({
//  user: '$MONGO_INITDB_ROOT_USERNAME',
//  pwd: '$MONGO_INITDB_ROOT_PASSWORD',
//  roles: [
//    { role: 'root', db: 'admin' },
//    { role: 'readAnyDatabase', db: "admin" },
//    { role: 'dbAdminAnyDatabase', db: "admin" },
//    { role: 'userAdminAnyDatabase', db: "admin" },
//    { role: 'readWrite', db: '$MONGO_INITDB_DATABASE' },
//    { role: 'readWrite', db: "admin" }
//  ]
//})
// db = db.getSiblingDB('$MONGO_INITDB_DATABASE');
// The second user is for the application startup, creating and maintaining indices and base data.
db.createUser({
  user: '$MONGO_INITDB_SETUP_USERNAME',
  pwd: '$MONGO_INITDB_SETUP_PASSWORD',
  roles: [
    { role: 'readAnyDatabase', db: "admin" },
    { role: 'dbAdminAnyDatabase', db: "admin" },
    { role: 'dbOwner', db: '$MONGO_INITDB_DATABASE' },
    { role: 'readWrite', db: '$MONGO_INITDB_DATABASE' }
  ]
})
// The third user is the main application user, which can only read and write to the app db
db.createUser({
  user: '$MONGO_INITDB_USERNAME',
  pwd: '$MONGO_INITDB_PASSWORD',
  roles: [
    { role: 'readAnyDatabase', db: "admin" },
    { role: 'readWrite', db: '$MONGO_INITDB_DATABASE' }
  ]
})
// db.createCollection("users")
// db.createCollection("fruits")
// db.createCollection("transfers")
EOF