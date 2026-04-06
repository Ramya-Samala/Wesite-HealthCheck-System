# Health Check System

A Go server that monitors websites by sending HTTP requests on a schedule and recording whether they are up or down. No external libraries used, only the Go standard library.


## Requirements

Go 1.21 or higher


## How to run

go run . --bind=127.0.0.1:8080 --checkfrequency=30s

If you want to run it with SSL:

go run . --bind=127.0.0.1:8443 --ssl --sslcert=cert.crt --sslkey=cert.key


## Available flags

--bind             address and port to listen on (default: 127.0.0.1:8080)
--checkfrequency   how often to run checks (default: 30s)
--ssl              enable HTTPS
--sslcert          path to SSL certificate file
--sslkey           path to SSL key file


## How it works

When the server starts it loads any previously saved checks from healthchecks.json. The background scheduler then runs all checks at the configured interval. Each check is run concurrently so it doesn't matter how many URLs you are monitoring, they all get checked at the same time. Results are saved to disk after every check so nothing is lost if the server restarts.


## API

### List all checks

GET /api/health/checks?page=1

Returns a paginated list of all health checks sorted alphabetically by endpoint. 10 results per page.

Response:
{
    "items": [
        {
            "id": "3b447fdf-d2e9-42bd-adcf-77d147b8b4dc",
            "status": "OK",
            "code": 200,
            "endpoint": "https://www.blizzard.com/en-us/",
            "checked": 1564065876,
            "duration": "127ms"
        },
        {
            "id": "a85113c0-c5e4-4657-ba88-9df8befdbaa1",
            "status": "Not Found",
            "code": 404,
            "endpoint": "https://www.blizzard.com/en-us/gotest",
            "checked": 1564065876,
            "duration": "127ms"
        },
        {
            "id": "1162d149-bd6f-4979-b9e5-7ee6f94f450a",
            "status": "Error",
            "code": 0,
            "endpoint": "https://gotest.blizzard.com",
            "error": "could not resolve host",
            "checked": 1564065876,
            "duration": "127ms"
        }
    ],
    "page": 1,
    "total": 3,
    "size": 10
}


### Get a single check

GET /api/health/checks/{id}

Returns one health check by id. Returns 404 if it doesn't exist.

Response:
{
    "id": "3b447fdf-d2e9-42bd-adcf-77d147b8b4dc",
    "status": "OK",
    "code": 200,
    "endpoint": "https://www.blizzard.com/en-us/",
    "checked": 1564065876,
    "duration": "127ms"
}


### Create a check

POST /api/health/checks

Body:
{
    "endpoint": "https://www.blizzard.com/en-us/"
}

Response:
{
    "id": "94a1d1e8-6e44-409e-9cb4-7bfcac2de1ae",
    "endpoint": "https://www.blizzard.com/en-us/"
}

Errors:
- 400 if endpoint is blank
- 400 if endpoint is not a valid URL
- 409 if a check for that URL already exists
- 500 if saving to disk fails


### Trigger a check manually

POST /api/health/checks/{id}/try?timeout=10s

Runs the check immediately instead of waiting for the next scheduled run.

Response when site is up:
{
    "id": "94a1d1e8-6e44-409e-9cb4-7bfcac2de1ae",
    "status": "OK",
    "code": 200,
    "endpoint": "https://www.blizzard.com/en-us/",
    "checked": 1564065975,
    "duration": "127ms"
}

Response when site is down or unreachable:
{
    "id": "1162d149-bd6f-4979-b9e5-7ee6f94f450a",
    "status": "Error",
    "code": 0,
    "endpoint": "https://gotest.blizzard.com",
    "error": "could not resolve host",
    "checked": 1564065975,
    "duration": "127ms"
}

Errors:
- 400 if timeout value is invalid
- 404 if check doesn't exist


### Delete a check

DELETE /api/health/checks/{id}

Removes the check from memory and from the saved file. Returns 204 with no body.

Errors:
- 404 if check doesn't exist


## Project structure

main.go       - starts the server and reads command line flags
handlers.go   - handles all incoming API requests
checker.go    - sends the HTTP request and records the result
scheduler.go  - runs checks on a timer in the background
store.go      - saves and loads checks from healthchecks.json
models.go     - defines the data structures used across the project
go.mod        - go module file

healthchecks.json  is  created automatically when the server runs