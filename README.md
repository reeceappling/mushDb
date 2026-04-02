# mushDb
this is a group of docker containers I use to host hobbyist stuff
the codebase is MESSY
# Running Locally
```bash
# With force recreate
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build --force-recreate
# Without force recreate
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build
# Without rebuilding
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up
```
# TODO:
mongosh into the db to setup users
