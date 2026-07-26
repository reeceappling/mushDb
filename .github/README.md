# mushDb

This is a group of containers (a docker-compose project) I use to keep track and notes of various fungus-related things for my hobby of mycology. The codebase is likely still VERY messy.

- The API acts as both a proxy to the webserver as well as the main api to interact with the database and RFID edge-services.
- The Database is a mongodb database (single-node replica set so we have access to transactions).
  - The Database data is stored in its own volume.
  - Pictures are stored in their own volume.
- The Webserver is a next.js app that serves the frontend.
- The API is made public via a Cloudflare tunnel.

[webserver README](../web/README.md)

[api README](../api/README.md)
# Running Locally
## Requirements
<details>
  <summary>Hardware Requirements</summary>

- TODO: CPU requirements!
- TODO: RAM requirements!
- TODO: Storage requirements!
- TODO: Network requirements!
</details>

<details>
  <summary>Dependencies</summary>

- Typescript -  TODO: VERSION
- Go (if developing or compiling the API locally)
- React -  TODO: VERSION
- Next.js -  TODO: VERSION
- MongoDB (Container) TODO: VERSION
- TODO: ADD MY OTHER REPO
</details>

## How to Run
### Setup
#### Cloudflare
##### Cloudflare Api for DNS
TODO: how to setup cloudflare api for dns
##### Cloudflare Tunnel
TODO: how to create cloudflare tunnel
#### Google OAuth
TODO: how to setup google oauth stuff

#### Environment
See /env/example.env for an example of the environment variables needed to run the project.
- COMPOSE_PROJECT_NAME: a project name utilized by docker compose
- ENVIRONMENT: qual/dev/devl, cert, or prod
- MAIN_API_EXTERNAL_PROTOCOL: http or https. I suggest https if going through cloudflare
- MAIN_API_EXTERNAL_PORT: The port on which the user will be accessing the cloudflare tunnel, NOT the api
- MAIN_API_EXTERNAL_HOST: The host on which the user will be accessing the service
- CLOUDFLARE_API_TOKEN: The cloudflare api token used for dns management for the CLOUDFLARE_ZONE_ID. You can create this in the cloudflare dashboard under "My Profile" -> "API Tokens". You can use the "Edit Cloudflare Workers" template, but you will also need to add "Zone.Zone" and "Zone.DNS" permissions for the specific zone.
- CLOUDFLARE_EMAIL: The email associated with the cloudflare account that has access to the CLOUDFLARE_ZONE_ID
- CLOUDFLARE_ZONE_ID: The zone ID for the domain you want to manage with cloudflare. You can find this in the cloudflare dashboard under the "Overview" tab for your domain, in the "API" section.
- ADMIN_GMAIL: The gmail account you would like to use as the admin account for the webserver.
- MONGO_INITDB_ROOT_USERNAME: The root username for the database. This is used to create the initial users in the database. Also for backups
- MONGO_INITDB_ROOT_PASSWORD: The root password for the database. This is used to create the initial users in the database. Also for backups
- MONGO_INITDB_USERNAME: The username for the api to access the database typically.
- MONGO_INITDB_PASSWORD: The password for the api to access the database typically.
- MONGO_INITDB_DATABASE: The name of the database to create/use for the project.
- MONGO_INITDB_SETUP_USERNAME: The database username used for initial table and index setup.
- MONGO_INITDB_SETUP_PASSWORD: The database password used for initial table and index setup.
- RFID_SECRET: secret used to register new rfid readers/writers which will interact with the server and the user
- GOTHIC_SESSION_SECRET: secret used to encrypt the session cookie for the webserver
- GOOGLE_CLIENT_ID: the google client ID used for oauth authentication for the webserver and api
- GOOGLE_CLIENT_SECRET: the google client secret used for oauth authentication for the webserver and api
- CF_TUNNEL_TOKEN: Cloudflare tunnel token used to run the cloudflare tunnel for the api. You can get this from the cloudflare dashboard when you create a tunnel, or by running `cloudflared tunnel create <tunnel-name>` in the terminal with cloudflared installed.
- LOCAL_DB_DIR: the local directory where all db data will be stored
- LOCAL_IMAGES_DIR: the local directory where all pictures will be stored which are uploaded by authenticated users


#### Filepaths
Whichever paths you are using for LOCAL_DB_DIR and LOCAL_IMAGES_DIR, ensure they exist and have the correct permissions for the docker user to read/write in them. See the "Running the containers" section below for an example of how to do this.


#### Mongodb keys
ensure [etc/mongodb.key](../etc/mongodb.key) exists. It is required to keep mongodb secure. If you don't have one, generate it with:
```bash
mkdir -p /etc
openssl rand -base64 741 > etc/mongodb.key
chmod 666 etc/mongodb.key
chown mongodb:mongodb etc/mongodb.key
```
### Running the containers
**The quick way** (If everything is already set up)
```bash
./scripts/run_main.sh
```
**The first way if the quick way is unavailable**

Before running the containers, ensure the api binary is built. It should exist at [/bin/mushApi](../bin/mushApi).
If it does not exist, run the following from the root of the project:
```bash
# Remove if exists already
rm -rf bin/mushApi
# Build the api binary
go build -o bin/mushApi .
```
You must also ensure the directories in which your pictures and database will be stored exist and have permissions for the docker user to read/write in them.
For the db, I use 
```txt
LOCAL_IMAGES_DIR="/opt/mush/data/pictures"
LOCAL_DB_DIR="/opt/mush/data/db"
```
So I have made sure that those two exist and have the right permissions via the following, or view [scripts/setup_dirs.sh](../scripts/setup_dirs.sh) for an example of how to do this between different environments.
```bash
MNGO_DIR="/opt/mush/data/db"
PICS_DIR="/opt/mush/data/pictures"
prep_dir() {
    local filepath="$1"
    echo "setting up path: $filepath"
    # Create the dir
    mkdir -p $filepath
    # Set the ownership to the current user and group
    chown -R $(id -u):$(id -g) $filepath
    # Allow all to read and write
    chmod -R a+rw $filepath
}

prep_dir $MNGO_DIR
prep_dir $PICS_DIR
```

#### With force recreate
```bash
MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build --force-recreate
```
#### Without force recreate
```bash

MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose --env-file env/.env.devhttps up --build
```
#### Without rebuilding
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
## Building Locally
Most of the building is done when you run docker compose for the project, but you can also build each part separately for development purposes.
### API
TODO: THIS
#### Regenerating Go Files

```bash
# Use subshell to change into the directory and run the script, so that we dont have to change back to the original directory after
(cd api/goGenerator
./buildAndGenerate.sh)
```
### Webserver
TODO: THIS

### Checking lines of code (because why not?)
Install cloc
```bash
  brew install cloc
```
Run cloc
```bash
cloc . --exclude-dir=node_modules,.github,.noCommit,.next --exclude-list-file=cypress.config.ts,eslint.config.mjs,next-env.d.ts,next.config.ts,package-lock.json,postcss.config.mjs,tailwind.config.ts,tsconfig.json,LICENSE,CODEOWNERS --exclude_ext=ico,svg,jpg,lock,png,env
```

# Misc. Repo Links
- [Acknowledgements](ACKNOWLEDGEMENTS.md) Still a template
- [Authors](AUTHORS.md)
- [Changelog](CHANGELOG.md)
- [Code of Conduct](CODE_OF_CONDUCT.md) Still a template
- [CODEOWNERS](CODEOWNERS)
- [Contributing](CONTRIBUTING.md)
- [Contributors](CONTRIBUTORS.md)
- [Funding](FUNDING.md) Still a template
- [Issue Template](ISSUE_TEMPLATE.md) Still a template
- [LICENSE](LICENSE) Still a template
- [Pull Request Template](pull_request_template.md) Still a template
- [Security](SECURITY.md) Still a template
- [Support](SUPPORT.md) Still a template
- [TODO](TODO.md) It's a lot...
