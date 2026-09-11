# EBANX Software Engineer Take-home Assignment

A small Go HTTP API for the EBANX take-home assignment. The implementation is
being developed incrementally so that every main commit represents a valid,
testable state of the project.

## Requirements

- Go 1.23 or newer
- An ngrok account and auth token only when exposing the API publicly

## Run locally

```sh
go run ./cmd/api
```

The server listens on port `8085` by default. Use `PORT` to select another port:

```sh
PORT=8080 go run ./cmd/api
```

Verify that the server is running:

```sh
curl -i http://localhost:8085/health
```

The root path (`/`) is not part of the assignment API and intentionally returns
`404 Not Found`. Use `/health` for the local availability check.

## Run with ngrok

The application uses the ngrok Go SDK already included in the project. Create
your local environment file from the committed example:

```sh
cp .env.example .env
```

Set your token in `.env`:

```dotenv
NGROK_AUTHTOKEN=<token>
```

Then start the application:

```sh
go run ./cmd/api
```

The local HTTP server starts before the ngrok connection is established, and the
public URL is printed in the application logs as soon as the tunnel is ready.
Startup logs also show the agent connection, authentication, endpoint creation
and heartbeat events without exposing the authentication token.
The `.env` file is ignored by Git and must not be committed. Variables already
exported by the operating system take precedence over values from the file.

## Verification

```sh
gofmt -w .
go test ./...
go vet ./...
```

## Postman

Import these files into Postman:

- `postman/EBANX Account API.postman_collection.json`
- `postman/Local.postman_environment.json`
- `postman/Ngrok.postman_environment.json`

Select **EBANX Account API - Local** to call the application at
`http://localhost:8085`. To test the public tunnel, edit `base_url` in the
**EBANX Account API - Ngrok** environment using the URL printed during startup,
without a trailing slash.

The collection contains a health check, the complete official EBANX workflow
and additional validation errors. Run the **Official workflow** folder in its
defined order because each request verifies the state produced by the previous
one, beginning with `POST /reset`. Every request includes Postman assertions for
its expected status and response body.

Run the official API workflow test independently with:

```sh
go test ./internal/httpapi -run TestOfficialAPIWorkflow -v
```

## Architecture

The API will use a simple layered architecture:

- The HTTP transport will handle routing, request parsing and HTTP responses.
- The account service will contain the business rules.
- The in-memory store will own the application state and its synchronization.
- The application bootstrap will create and connect these components.

This structure keeps business rules independent from HTTP and makes them easy to
test, while avoiding abstractions that are unnecessary for the size of the
challenge. Business logic directly in handlers was rejected because it couples
the rules to HTTP. A full Clean or Hexagonal Architecture was also rejected
because its additional ports, adapters and mappings would not provide
proportional value here.

The account service owns a small `Store` interface containing only the operations
required by its business rules. This keeps the domain independent from the
in-memory implementation without introducing repository abstractions or mapping
layers that are unnecessary for the challenge.

## Scope and current status

Durability is intentionally not implemented because it is explicitly outside the
assignment scope. Application state will exist only during the process lifetime
and will be cleared through `POST /reset` once the account operations are added.

Account IDs are strings, matching the API contract. Balances are represented by
`int64`, which avoids floating-point rounding and is sufficient for the integer
amounts defined by the assignment.

Phase 8 adds an integration test that reproduces the complete official EBANX API
workflow against the real HTTP handler, account service and in-memory store. The
test validates each response in order and confirms that the final failed
transfer preserves both account balances.
