# mushDb

This is a group of containers (a docker-compose project) I use to keep track and notes of various fungus-related things for my hobby of mycology. The codebase is likely still VERY messy.

- The API acts as both a proxy to the webserver as well as the main api to interact with the database and RFID edge-services.
- The Database is a mongodb database (single-node replica set so we have access to transactions).
  - The Database data is stored in its own volume.
  - Pictures are stored in their own volume.
- The Webserver is a next.js app that serves the frontend.
- The API is made public via a Cloudflare tunnel.
## TODO
<details>
  <summary>TODOs</summary>

- Backend Testing
- Frontend Testing (Cypress?)
- Add metrics and monitoring (grafana?)
- Add DB and image backup system
- Dictation on the frontend for hands-free operation when working in a sterile environment.
</details>


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
### Environment setup
TODO: how to setup your env file
### Running the containers
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
(cd rfid/goGenerator
./buildAndGenerate.sh)
```
### Webserver
TODO: THIS


# Repo Links
- [Acknowledgements](ACKNOWLEDGEMENTS.md) Still a template
- [Authors](AUTHORS.md)
- [Changelog](CHANGELOG.md)
- [CODEOWNERS](CODEOWNERS)
- [Contributing](CONTRIBUTING.md)
- [Contributors](CONTRIBUTORS.md)
- [Funding](FUNDING.md) Still a template
- [Issue Template](ISSUE_TEMPLATE.md) Still a template
- [Pull Request Template](pull_request_template.md) Still a template
- [Security](SECURITY.md) Still a template
