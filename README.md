# EBANX Software Engineer Take-home Assignment

A small Go HTTP API for the EBANX take-home assignment. The implementation is
being developed incrementally so that every main commit represents a valid,
testable state of the project.

## Requirements

- Go 1.23 or newer
- An ngrok account and auth token only when exposing the API publicly

## Run locally

```sh
go run .
```

The server listens on port `8085` by default. Use `PORT` to select another port:

```sh
PORT=8080 go run .
```

Verify that the server is running:

```sh
curl -i http://localhost:8085/health
```

## Run with ngrok

The application uses the ngrok Go SDK already included in the project. Set the
auth token and start the application:

```sh
NGROK_AUTHTOKEN=<token> go run .
```

The public URL is printed in the application logs. The token must be supplied
through the environment and must not be committed to the repository.

## Verification

```sh
gofmt -w .
go test ./...
go vet ./...
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

Interfaces will not be introduced until there is more than one implementation or
a concrete testability need. This keeps the code malleable without anticipating
requirements that are outside the specification.

## Scope and current status

Durability is intentionally not implemented because it is explicitly outside the
assignment scope. Application state will exist only during the process lifetime
and will be cleared through `POST /reset` once the account operations are added.

Phase 0 provides the executable HTTP bootstrap, configurable port, optional
ngrok forwarding and a health check. Account operations are intentionally left
for the following implementation phases.
