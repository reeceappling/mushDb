let dbName = process.env.MONGO_INITDB_DATABASE
db = db.getSiblingDB('admin');
// move to the admin db - always created in Mongo
// log as root admin if you decided to authenticate in your docker-compose file...
// create and move to your new database
try {
    print("creating user")
    db.createUser(
        {
            user: process.env.MONGO_INITDB_ROOT_USERNAME,
            pwd: process.env.MONGO_INITDB_ROOT_PASSWORD,
            roles: [ { role: "readAnyDatabase", db: "admin" }, { role: "dbAdminAnyDatabase", db: "admin" }, { role: "userAdminAnyDatabase", db: "admin" }, { role: "readWrite", db: dbName }]
        }
    );
    db.log.insertOne({"message": "User created ."});
} catch (error) {
    print(`Failed creating developer db user:\n${error}`);
}
try {
    print("creating user")
    db.createUser(
        {
            user: process.env.MONGO_INITDB_USERNAME,
            pwd: process.env.MONGO_INITDB_PASSWORD,
            roles: [ { role: "readAnyDatabase", db: "admin" }, { role: "dbAdminAnyDatabase", db: "admin" }, { role: "userAdminAnyDatabase", db: "admin" }, { role: "readWrite", db: dbName }]
        }
    );
    db.log.insertOne({"message": "User created ."});
} catch (error) {
    print(`Failed creating developer db user:\n${error}`);
}
// add new collection // TODO: this

db = db.getSiblingDB(dbName)