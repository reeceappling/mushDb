# mushDb

This is a group of containers (a docker-compose project) I use to keep track and notes of various fungus-related things for my hobby of mycology. The codebase is likely still VERY messy.

- The API acts as both a proxy to the webserver as well as the main api to interact with the database and RFID edge-services.
- The Database is a mongodb database (single-node replica set so we have access to transactions).
  - The Database data is stored in its own volume.
  - Pictures are stored in their own volume.
- The Webserver is a next.js app that serves the frontend.
- The API is made public via a Cloudflare tunnel.
## TODO
- Backend Testing
- Frontend Testing (Cypress?)
- Add metrics and monitoring (grafana?)
- Add DB and image backup system
- Dictation on the frontend for hands-free operation when working in a sterile environment.

# Running Locally
### With force recreate
```bash
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build --force-recreate
```
### Without force recreate
```bash

MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build
```
### Without rebuilding
```bash
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up
```

## Viewing Logs
For each service, only the last 50 lines
```bash
docker compose --env-file env/.env.devhttps logs --tail 50 api
docker compose logs --tail 50 web
docker compose logs --tail 50 mushdb
```
following the api logs
```bash
docker compose logs -f  api
```

## Regenerating Go Files

```bash
# Use subshell to change into the directory and run the script, so that we dont have to change back to the original directory after
(cd rfid/goGenerator
./buildAndGenerate.sh)
```