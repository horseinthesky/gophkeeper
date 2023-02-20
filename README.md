# 🔒 gophkeeper

**gophkeeper** is a PoC secret storage service.

![main](https://github.com/horseinthesky/gophkeeper/blob/diploma/media/main.png)

## ✨ Features

- 📦 Manage all secrets with a nice [bubbletea powered](https://github.com/charmbracelet/bubbletea) UI
- 🚀 Fast secret management thanks to automatic caching
- 💾 Transparent background synchronization with the server
- 💪 Async execution for improved performance

### New secret

Select a secret kind

![new](https://github.com/horseinthesky/gophkeeper/blob/diploma/media/new.png)

Fill new secret form

![entry](https://github.com/horseinthesky/gophkeeper/blob/diploma/media/entry.png)

Display your secret info

![show](https://github.com/horseinthesky/gophkeeper/blob/diploma/media/show.png)

### Supported secret kinds

- Login/Password pairs
- Arbitrary test
- Arbitrary bytes (files)
- Bank card credentials

## ⚡️ Requirements

- Git
- Docker
- Go >= 1.19

## 📦 Installation

Clone the repo.

Next install dependencies

```bash
make init
```

Then build packages

```bash
make build
```

## 🚀 Usage

### Server

The Server supports the following settings:

- `env` - environment determines what the logging level and log format will be
  - `dev` - plain text colored `INFO` level logs
  - `prod` (**default**) - JSON `WARN` level logs
- `address` - `address:port` to listen on (defaults to `localhost:8080`)
- `dsn` - PostgreSQL database DSN
- `clean` - database cleanup time interval (defaults to `15m`)

All can set all the settings in the config file (`-c` flag) or via env vars (overrides config file values) with the same names prefixed with `GOPHKEEPER_` (e.g. `GOPHKEEPER_ENV`).

Run server with:
```
./gs -c <your_server_config.yml>
```

### Client

Client settings are the following:

- `user` (**mandatory**) - your username
- `password` (**mandatory**) - your password
- `encrypt` (default is `true`) - if gophkeeper should encrypt your secrets
- `key` (**mandatory** 32 bytes master password if `encrypt` set to `true`) - this key will be used to encrypt your secrets
- `env` - environment determines what the logging level and log format will be
  - `dev` - plain text colored `INFO` level logs
  - `prod` (**default**) - JSON `WARN` level logs
- `address` - `address:port` of the server to connect to (defaults to`localhost:8080`)
- `dsn` - PostgreSQL database DSN
- `sync` - secret synchronization time interval (defaults to `15s`)
- `clean` - database cleanup time interval (defaults to `1m`)

All can set all the settings in the config file (`-c` flag) or via env vars (overrides config file values) with the same names prefixed with `GOPHKEEPER_` (e.g. `GOPHKEEPER_ENV`).

Run client with:
```
./gc -c <your_client_config.yml>
```

The client will **automatically** register/login (if you are an existing user) with provided credentials.

## 🔨 Dev

For development you will need additional tools:

- [sqlc](https://github.com/kyleconroy/sqlc)
- [go mirgate](https://github.com/golang-migrate/migrate)
- [gomock](https://github.com/golang/mock)

Install them with

```bash
make dev
```

Next prepare test databases for client/server

```bash
make mkdb
make migrateup
```

You can refresh (purge and reinstall) your DBs with

```bash
make refreshdb
```

You an also renew client/server certificates with

```bash
make cert
```
