# 🔒 gophkeeper

**gophkeeper** is a PoC secret storage service.

![main](https://github.com/horseinthesky/gophkeeper/blob/diploma/main.jpg)

## ✨ Features

- 📦 Manage all secrets with a nice [bubbletea powered](https://github.com/charmbracelet/bubbletea) UI
- 🚀 Fast secret management thanks to automatic caching
- 💾 Transparent background synchronization with the server
- 💪 Async execution for improved performance

### New secret

Select a secret kind

![new](https://github.com/horseinthesky/gophkeeper/blob/diploma/new.jpg)

Fill new secret form

![entry](https://github.com/horseinthesky/gophkeeper/blob/diploma/entry.jpg)

Display your secret info

![show](https://github.com/horseinthesky/gophkeeper/blob/diploma/show.jpg)

### Supported secret kinds

- Login/Password pairs
- Arbitrary test
- Arbitrary bytes (files)
- Bank card credentials

## ⚡️ Requirements

- Git
- Docker
- Go  >= 1.19

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

### 🔨 Dev

For development you will need additional tools:
- [sqlc](https://github.com/kyleconroy/sqlc)
- [go mirgate](https://github.com/golang-migrate/migrate)

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
