# EBANX Software Engineer Take-home Assignment

A small Go HTTP API for the EBANX take-home assignment. It implements account
deposits, withdrawals, transfers and balance queries using concurrency-safe
in-memory state.

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
Startup logs also show the agent connection, authentication and endpoint
creation without exposing the authentication token.
The `.env` file is ignored by Git and must not be committed. Variables already
exported by the operating system take precedence over values from the file.

## Verification

```sh
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

## API contract

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/reset` | Clear all in-memory account state |
| `GET` | `/balance?account_id={id}` | Return an account balance |
| `POST` | `/event` | Process a deposit, withdrawal or transfer |
| `GET` | `/health` | Check local application availability |

Event examples:

```json
{"type":"deposit","destination":"100","amount":10}
{"type":"withdraw","origin":"100","amount":5}
{"type":"transfer","origin":"100","destination":"300","amount":15}
```

The assignment responses are preserved exactly: successful events return
`201`, missing accounts return `404 0`, balances return their numeric value and
reset returns `200 OK` with `OK` in the response body.

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

The API uses a simple layered architecture:

- The HTTP transport handles routing, request parsing and HTTP responses.
- The account service contains the business rules and coordinates operations.
- The in-memory store owns the application state and its synchronization.
- The application bootstrap creates and connects these components.

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

## Design decisions and trade-offs

Durability is intentionally not implemented because it is explicitly outside the
assignment scope. Application state will exist only during the process lifetime
and can be cleared through `POST /reset`.

Account IDs are strings, matching the API contract. Balances are represented by
`int64`, which avoids floating-point rounding and is sufficient for the integer
amounts defined by the assignment.

The service serializes state-changing operations with a mutex. `SaveAll` writes
both sides of a transfer while holding the store lock, so concurrent requests
cannot observe or leave a partially applied transfer. Balance queries use read
locks and never create or update accounts.

The project deliberately avoids a web framework, database, dependency-injection
container and additional mapping layers. The standard library and one small
store interface keep the implementation easy to read, test and modify for the
scope of the exercise.

For a production system, the in-memory store would be replaced by transactional
durable storage. Multiple application instances would require database-level
concurrency control or another distributed coordination strategy. Production
hardening would also include authentication, rate limiting, request and server
timeouts, graceful shutdown, structured observability and explicit monetary
limits. These concerns are documented but intentionally not implemented because
they are outside the assignment requirements.

## Test strategy

Unit tests exercise deposits, withdrawals, transfers, resets, error paths and
concurrent state changes against the real in-memory store. HTTP tests validate
transport mapping and exact response bodies. The full workflow test runs the
official EBANX sequence through the real HTTP handler, service and store, then
confirms that a failed transfer leaves both account balances unchanged.
