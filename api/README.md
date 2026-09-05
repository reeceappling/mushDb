# Go API
Entrypoint for all client requests to the project. Covers the api for interacting with the database, rfid readers, file retrievals, and even acts as a proxy to the webserver.
### TODO: THIS!-----------------------------------------------------
## Main area
### TODO: THIS!-----------------------------------------------------
## Directories
### cache
Package for caching data in memory, on disk, and in external services. (NOT UTILIZED YET)
### env
Package for utilizing environment variables and secrets.
### goGenerator
Standalone go template program to generate go code from templates. Used to generate the api and database code for different entry types.
### gotel
package for Go Otel
### mocks
package for generated mocks
### pics
Package for storing and retrieving pictures from the filesystem utilizing context.
### request
Request IDs, traces, and time storage on the context of each request.