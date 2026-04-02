# mushDb
this is a group of docker containers I use to host hobbyist stuff
the codebase is MESSY
# Running Locally
With a force-recreate
```bash
# Without force recreate
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.dev up --build --force-recreate
# With force recreate
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttp up --build --force-recreate
```
# TODO:
mongosh into the db to setup users
