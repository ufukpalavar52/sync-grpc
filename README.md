# imapsync-grpc

A gRPC service that manages the user accounts, IMAP sync jobs, credential encryption, log
streaming and statistics behind an IMAP mailbox migration ("imapsync") platform.

The service is written in Go, stores its data in PostgreSQL through GORM, validates every
request with [protovalidate](https://github.com/bufbuild/protovalidate), and writes
structured JSON logs with zap + lumberjack.

---

## Table of contents

- [Architecture](#architecture)
- [gRPC API](#grpc-api)
  - [UserService](#userservice-userv1)
  - [SyncService](#syncservice-syncv1)
  - [EncryptService](#encryptservice-encryptv1)
  - [SyncLogService](#synclogservice-logv1)
  - [AnalysisService](#analysisservice-analysisv1)
- [Error codes](#error-codes)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [Database schema](#database-schema)
- [Regenerating the protobuf code](#regenerating-the-protobuf-code)
- [Logging](#logging)
- [Tests](#tests)
- [Project layout](#project-layout)

---

## Architecture

```
proto/api/v1/*.proto        ──buf generate──▶  pkg/pb/**       (generated gRPC stubs)

cmd/server/main.go          bootstraps the process: DB pool, encryption key, gRPC server
      │
      ├── cmd/interceptor   unary + stream logging interceptors
      └── cmd/handler       gRPC servers: validate request → call service → map to protobuf
                │
                └── internal/service   business logic (users, syncs, encryption, log tail, stats)
                          │
                          └── internal/repository   GORM data access (PostgreSQL)
                                    │
                                    └── internal/model   entities, search filters, pagination
```

Cross-cutting packages:

| Package            | Responsibility                                                        |
| ------------------ | --------------------------------------------------------------------- |
| `config`           | Environment-driven configuration and the GORM/PostgreSQL connection    |
| `logger`           | zap logger with console + rotating JSON file output                    |
| `internal/util`    | AES-GCM encryption, bcrypt hashing, gRPC status errors, metadata, JSON |
| `internal/global`  | Process-wide state (the active encryption key)                         |

Notable behaviours:

- **Startup encryption key.** On boot the process looks up the encryption key for
  `EncryptionKeyVersion` in the `encryption_keys` table. If it does not exist, a random
  32-byte AES key is generated and stored. The key is kept in `global.EncKey` and used to
  encrypt sync credentials. Startup fails hard if this step fails.
- **Server reflection** is registered, so `grpcurl`, Postman and similar tools can discover
  the API without a local copy of the `.proto` files.
- **Passwords** of platform users are hashed with bcrypt (cost 12); IMAP account passwords
  and OAuth client secrets are encrypted with AES-256-GCM.

---

## gRPC API

The server listens on `APP_PORT` (default `50051`) over plaintext TCP — put it behind a
gateway/mesh if you need TLS.

Every request may carry a `Transaction-Id` metadata header; it is picked up by the logging
interceptor and attached to all log lines for that call.

### UserService (`user.v1`)

| RPC                  | Request                     | Response       | Notes                                                              |
| -------------------- | --------------------------- | -------------- | ------------------------------------------------------------------ |
| `CreateUser`         | `CreateUserRequest`         | `UserResponse` | Fails with `ALREADY_EXISTS` if the e-mail is taken. Hashes password |
| `UpdateUser`         | `UpdateUserRequest`         | `UserResponse` | Updates name / profile image / e-mail verification time only        |
| `UpdateUserPassword` | `UpdateUserPasswordRequest` | `UserResponse` | Requires the current password (re-authenticates first)             |
| `ResetUserPassword`  | `ResetUserPasswordRequest`  | `UserResponse` | Sets a new password without knowing the old one                    |
| `AuthUser`           | `AuthUserRequest`           | `UserResponse` | Credential check; `NOT_FOUND` on a wrong e-mail *or* password       |
| `GetUser`            | `GetUserRequest`            | `UserResponse` | Lookup by e-mail                                                    |

Validation: `email` must be a valid address of at most 255 characters, `password`,
`newPassword` and `name` must be at least 4 characters.

`UserResponse` never contains the password hash. `createdAt` / `updatedAt` are formatted as
`YYYY-MM-DD HH:MM:SS`.

### SyncService (`sync.v1`)

| RPC                 | Request                 | Response                 | Notes                                                    |
| ------------------- | ----------------------- | ------------------------ | -------------------------------------------------------- |
| `CreateSync`        | `ImapSyncRequest`       | `ImapSyncResponse`       | Creates a sync job in `PENDING` status                    |
| `BulkCreateSync`    | `BulkCreateSyncRequest` | `BulkCreateSyncResponse` | Per-item result; failures are returned, not thrown        |
| `GetSync`           | `GetSyncRequest`        | `ImapSyncResponse`       | Lookup by `transactionId`, credentials omitted            |
| `GetSyncFullDetail` | `GetSyncRequest`        | `ImapSyncFullResponse`   | Same lookup including the (encrypted) credential fields   |
| `ListSync`          | `ListSyncRequest`       | `ListSyncResponse`       | Filtering + pagination                                    |
| `UpdateStatus`      | `UpdateStatusRequest`   | `ImapSyncResponse`       | Also stamps the elapsed time when the job reaches an end  |

**Creating a sync.** `CreateSync` verifies that `userId` exists and that `transactionId` is
not already used, then persists the job:

- If `isEncrypted` is unset or `false`, the service encrypts `sourcePassword`,
  `destPassword`, `sourceClientSecret` and `destClientSecret` with the active AES key before
  storing them. Set `isEncrypted: true` when the caller has already encrypted those values
  (for example via `EncryptService`) so they are stored verbatim.
- The key version in use is stored on the row (`encryption_version`), so credentials stay
  decryptable after a key rotation.
- Status is always forced to `PENDING`; the `status` field in the request is validated but
  not used for the initial row.

`BulkCreateSync` runs the same logic per element and returns two lists: `syncs` for the rows
that were created and `errors` (`{transactionId, message}`) for the ones that were not, so a
partially successful batch still returns `OK`.

**Job statuses** (`internal/util/const.go`): `PENDING`, `IN_PROGRESS`, `FAILED`,
`COMPLETED`, `CANCELED`. When `UpdateStatus` moves a job to anything other than `PENDING` or
`IN_PROGRESS`, `finish_time` is set to the number of seconds since the row was created and
returned as `FinishedTime`.

**Listing.** `ListSyncRequest` requires `userId` (> 0). `transactionId` and `status` are
matched exactly; `sourceHost`, `sourceUser`, `destHost` and `destUser` are matched with
`LIKE %value%`. Empty fields are ignored. Results are ordered by `id` descending.
`limit`/`offset` are optional — note that `total` is only computed when at least one of them
is greater than zero; an unpaginated call returns every matching row with `total = 0`.

### EncryptService (`encrypt.v1`)

| RPC       | Request          | Response          |
| --------- | ---------------- | ----------------- |
| `Encrypt` | `EncryptRequest` | `EncryptResponse` |
| `Decrypt` | `EncryptRequest` | `EncryptResponse` |

AES-256-GCM helper endpoints. `version` selects the key: when omitted the active
`EncryptionKeyVersion` is used, otherwise the key with that version is loaded from the
database (`NOT_FOUND` if it does not exist). The response echoes the version that was used.
Ciphertext is base64 of `nonce || ciphertext`.

> These RPCs are an unauthenticated encryption oracle for anything that can reach the port —
> keep the service on a private network.

### SyncLogService (`log.v1`)

| RPC          | Request          | Response                |
| ------------ | ---------------- | ----------------------- |
| `StreamLogs` | `SyncLogRequest` | `stream SyncLogChunk`   |

Server-streaming tail of `${SYNC_LOG_PATH}/${transactionId}.log`. The file must already
exist, otherwise the RPC fails. Lines are buffered and flushed to the client either every
200 ms or as soon as 100 lines have accumulated.

The stream ends when:

- the log contains `EX_OK` (imapsync finished successfully) — 5 seconds later;
- the log contains an `EXIT_..._FAILURE_` marker while the job is still reported as running
  — 50 seconds later;
- no new data arrives for 500 seconds (10 seconds if `status` in the request is already a
  terminal one);
- the client cancels the context.

Pass the current job `status` in the request so the server can pick the short idle timeout
for jobs that are already finished.

### AnalysisService (`analysis.v1`)

| RPC                 | Request             | Response                     |
| ------------------- | ------------------- | ---------------------------- |
| `GetSyncCountStats` | `SyncsStatsRequest` | `SyncsStatsResponseByStatus`  |

Returns one `{status, count, avgTime}` entry per status, plus a synthetic `Total` entry
whose `avgTime` is the count-weighted average across all statuses. `avgTime` is the average
`finish_time` in seconds. Passing `userId` scopes the stats to a single user; omitting it
(or sending `0`) returns platform-wide numbers.

---

## Error codes

| Code               | When                                                                        |
| ------------------ | --------------------------------------------------------------------------- |
| `INVALID_ARGUMENT` | protovalidate rejected the request; the message contains the failing rules   |
| `NOT_FOUND`        | user not found, sync not found, key version not found, wrong e-mail/password |
| `ALREADY_EXISTS`   | e-mail already registered, or `transactionId` already used                   |
| `UNKNOWN`          | database or crypto failure — the underlying error text is passed through     |

---

## Getting started

### Requirements

- Go 1.25+
- PostgreSQL
- [`buf`](https://buf.build/docs/installation) (only if you change the `.proto` files)

### Run

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=imapsync
export DB_PASSWORD=secret
export DB_DATABASE=imapsync
export SYNC_LOG_PATH=/var/log/imapsync

go mod download
go run ./cmd/server
```

You should see `gRPC server listening at [::]:50051`.

### Build

```bash
go build -o bin/imapsync-grpc ./cmd/server
```

### Try it out

Reflection is enabled, so `grpcurl` works without `.proto` files:

```bash
grpcurl -plaintext localhost:50051 list

grpcurl -plaintext -d '{"email":"jane@example.com","password":"s3cret","name":"Jane"}' \
  localhost:50051 user.v1.UserService/CreateUser

grpcurl -plaintext -H 'Transaction-Id: 8f3c1b' -d '{
  "transactionId":"8f3c1b",
  "userId":1,
  "sourceHost":"imap.old.example.com","sourceUser":"jane@old.example.com","sourcePassword":"pass",
  "destHost":"imap.new.example.com","destUser":"jane@new.example.com","destPassword":"pass",
  "status":"PENDING"
}' localhost:50051 sync.v1.SyncService/CreateSync

grpcurl -plaintext -d '{"userId":1,"limit":20,"offset":0}' \
  localhost:50051 sync.v1.SyncService/ListSync

grpcurl -plaintext -d '{"transactionId":"8f3c1b","status":"IN_PROGRESS"}' \
  localhost:50051 log.v1.SyncLogService/StreamLogs
```

---

## Configuration

All settings come from environment variables (see `config/config.go`); every one has a
default, so the service starts with none of them set — but the database ones must be
supplied for a real deployment.

| Variable                | Default            | Description                                        |
| ----------------------- | ------------------ | -------------------------------------------------- |
| `APP_NAME`              | `imap-sync`        | Added to every log line                            |
| `APP_ENV`               | `dev`              | Added to every log line                            |
| `APP_PORT`              | `50051`            | gRPC listen port                                   |
| `DB_HOST`               | `localhost`        | PostgreSQL host                                    |
| `DB_PORT`               | `5432`             | PostgreSQL port                                    |
| `DB_USER`               | *(empty)*          | PostgreSQL user                                    |
| `DB_PASSWORD`           | *(empty)*          | PostgreSQL password                                |
| `DB_DATABASE`           | *(empty)*          | PostgreSQL database name                           |
| `DB_TIMEZONE`           | `Europe/Istanbul`  | Session time zone                                  |
| `DB_SSL_MODE`           | `0`                | `1` enables SSL, anything else disables it         |
| `DB_POOL_MAX_IDLE_CONN` | `10`               | Idle connections in the pool                       |
| `DB_POOL_MAX_OPEN_CONN` | `100`              | Maximum open connections (lifetime is 1 hour)      |
| `EncryptionKeyVersion`  | `v1`               | Active AES key version (note the mixed-case name)  |
| `SYNC_LOG_PATH`         | *(empty)*          | Directory holding `<transactionId>.log` files      |
| `LOG_PATH`              | `logs/app.log`     | Application log file                               |
| `LOG_LEVEL`             | `info`             | `debug`, `info` or `error`                         |

---

## Database schema

The service does **not** run migrations — create the tables before starting it. The GORM
entities in `internal/model/entity.go` expect the following shape:

```sql
CREATE TABLE users (
    id                  BIGSERIAL PRIMARY KEY,
    name                TEXT,
    email               TEXT,
    password            TEXT,
    profile_image       TEXT,
    email_verified_time BIGINT,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ
);

CREATE TABLE imap_syncs (
    id                   BIGSERIAL PRIMARY KEY,
    transaction_id       TEXT UNIQUE,
    source_host          TEXT,
    source_user          TEXT,
    source_auth_user     TEXT,
    source_password      TEXT,
    source_ssl           BOOLEAN,
    source_tenant_id     TEXT,
    source_client_id     TEXT,
    source_client_secret TEXT,
    source_port          INTEGER,
    dest_host            TEXT,
    dest_user            TEXT,
    dest_auth_user       TEXT,
    dest_password        TEXT,
    dest_ssl             BOOLEAN,
    dest_tenant_id       TEXT,
    dest_client_id       TEXT,
    dest_client_secret   TEXT,
    dest_port            INTEGER,
    skip_header          JSONB,
    exclude_folders      JSONB,
    status               TEXT,
    user_id              BIGINT,
    encryption_version   TEXT,
    finish_time          DOUBLE PRECISION,
    created_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ
);

CREATE TABLE encryption_keys (
    id         BIGSERIAL PRIMARY KEY,
    key        TEXT,
    version    TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

`encryption_keys.key` holds the base64-encoded AES-256 key and is populated automatically on
first start. Treat that table as a secret.

---

## Regenerating the protobuf code

The `.proto` sources live in `proto/api/v1/` and the generated Go code in `pkg/pb/`.
Generation is driven by `buf` (`buf.work.yaml`, `proto/buf.yaml`, `buf.gen.yaml`) with the
remote `protocolbuffers/go` and `grpc/go` plugins:

```bash
buf lint
buf generate
```

`proto/` and `pkg/` are listed in `.gitignore`, so a fresh checkout may need `buf generate`
(and a `buf` dependency update for `buf.build/bufbuild/protovalidate`) before it compiles.

---

## Logging

`logger.InitLog` writes to two sinks at once:

- **stdout** — human-readable, colourised console output;
- **`LOG_PATH`** — JSON lines rotated by lumberjack (100 MB per file, 7 backups, 28 days,
  gzip-compressed).

Every line carries `app`, `env`, `hostname` and a `caller` field (`file.go:line - Type.Method()`).
The unary and stream interceptors log one entry per RPC with `method`, `code`, `duration`
(seconds), `clientIp` and the `Transaction-Id` metadata value, which makes it easy to trace a
single migration end to end.

---

## Tests

Unit tests live under `test/` and mirror the `internal/` tree. Repository tests run against
`go-sqlmock`, so no database is required.

```bash
go test ./...
```

---

## Project layout

```
cmd/
  server/       process entrypoint, gRPC server wiring
  handler/      gRPC service implementations (validation + protobuf mapping)
  interceptor/  unary and stream logging interceptors
config/         environment configuration and the GORM connection
internal/
  service/      business logic
  repository/   GORM data access
  model/        entities, search filters, generic pagination result
  util/         crypto, hashing, gRPC errors, metadata, JSON, constants
  global/       process-wide state (active encryption key)
logger/         zap + lumberjack setup
proto/api/v1/   protobuf definitions
pkg/pb/         generated gRPC/protobuf code
test/           unit tests
```
