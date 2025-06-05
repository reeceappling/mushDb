let dbName =process.env.MONGO_INITDB_DATABASE
db.myColl.insertMany([
    { name: "first 1", value: 10 }, // TODO: FIX!!!!
    { name: "second 2", value: 20 },
    // Add more default documents as needed
]);
db.createUser(
    {
        user: process.env.MONGO_INITDB_ROOT_USERNAME,
        pwd: process.env.MONGO_INITDB_ROOT_PASSWORD,
        //roles: [ { role: "readAnyDatabase", db: "admin" }, { role: "dbAdminAnyDatabase", db: "admin" }, { role: "userAdminAnyDatabase", db: "admin" }]
        roles: [ { role: "readWrite", db: dbName }]
    });
db = db.getSiblingDB('admin');
//db.auth(process.env.MONGO_INITDB_ROOT_USERNAME, process.env.MONGO_INITDB_ROOT_PASSWORD)

db.createUser(
    {
        user: process.env.MONGO_INITDB_USERNAME,
        pwd: process.env.MONGO_INITDB_PASSWORD,
        roles: [ { role: "readAnyDatabase", db: "admin" }, { role: "dbAdminAnyDatabase", db: "admin" }, { role: "userAdminAnyDatabase", db: "admin" }]
    });
db = db.getSiblingDB(dbName)
// db = db.getSiblingDB('admin');
// // move to the admin db - always created in Mongo
// //db.auth(fs.readFileSync('/run/secrets/db-root-username', 'utf-8'), fs.readFileSync('/run/secrets/db-root-password', 'utf-8'));
// // log as root admin if you decided to authenticate in your docker-compose file...
// // create and move to your new database
// try {
//     print("creating user")
//     db.createUser(
//         {
//             user: fs.readFileSync('/run/secrets/db-server-username', 'utf-8'),
//             pwd: fs.readFileSync('/run/secrets/db-server-pass', 'utf-8'),
//             // user: process.env.MONGO_USER,
//             // pwd: process.env.MONGO_PASS,
//             roles: [{ role: "readWrite", db: "test" }] // TODO: FIX DB NAME
//         }
//     );
// } catch (error) {
//     print(`Failed creating developer db user:\n${error}`);
// }
// //db.log.insertOne({"message": "User created ."});
// db.createCollection('collection_test');
// // add new collection