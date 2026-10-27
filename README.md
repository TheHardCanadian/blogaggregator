# Gator

Gator is a command-line RSS feed aggregator written in Go. You register a user, add RSS feeds, and follow the ones you're interested in. While a background aggregator is running, it fetches new posts into PostgreSQL, and you can browse them from your terminal.

## Prerequisites

You'll need these installed:

- **[Go](https://go.dev/dl/) 1.26 or later**: used to install the CLI.
- **[PostgreSQL](https://www.postgresql.org/download/) 13 or later**: stores users, feeds, and posts. It can run locally or anywhere you can reach over the network.
- **[goose](https://github.com/pressly/goose)**: creates the database schema. Install it with:

  ```sh
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

Check that Go and Postgres are available:

```sh
go version
psql --version
```

## Installation

Install the CLI with `go install`:

```sh
go install github.com/thehardcanadian/blogaggregator@latest
```

This puts a binary named `blogaggregator` in your Go bin directory (`$GOBIN` if set, otherwise `~/go/bin`). Make sure that directory is on your `PATH`.

The examples below call the binary `blogaggregator`. If you'd rather type `gator`, add an alias to your shell profile:

```sh
alias gator=blogaggregator
```

## Setup

### 1. Create a database

Connect to Postgres with `psql` and create a database for Gator:

```sql
CREATE DATABASE gator;
```

Your connection URL will look like this. Fill in your own username and password:

```
postgres://<user>:<password>@localhost:5432/gator?sslmode=disable
```

Keep `sslmode=disable` for a local server that doesn't have SSL set up, or the connection will fail.

### 2. Create the schema

The migrations are in [`sql/schema`](sql/schema). Clone the repo and run them with goose:

```sh
git clone https://github.com/thehardcanadian/blogaggregator.git
cd blogaggregator
goose -dir sql/schema postgres "postgres://<user>:<password>@localhost:5432/gator?sslmode=disable" up
```

### 3. Create the config file

Gator reads its settings from `~/.gatorconfig.json` and won't start without it. Create the file and set `db_url` to your connection URL:

```json
{
  "db_url": "postgres://<user>:<password>@localhost:5432/gator?sslmode=disable"
}
```

Gator adds a `current_user_name` field itself whenever you run `register` or `login`, so you don't need to set it.

## Usage

```sh
blogaggregator <command> [args...]
```

### Quick start

Register a user. This also logs you in:

```sh
blogaggregator register alice
```

Add a couple of feeds. You automatically follow any feed you add:

```sh
blogaggregator addfeed "Hacker News" https://news.ycombinator.com/rss
blogaggregator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml
```

Start the aggregator and leave it running. Each tick, it fetches the feed that was updated least recently and saves any new posts. Stop it with `Ctrl+C`:

```sh
blogaggregator agg 1m
```

In another terminal, read the latest posts from the feeds you follow:

```sh
blogaggregator browse 10
```

### Commands

| Command | What it does |
| --- | --- |
| `register <name>` | Create a user and log in as them. |
| `login <name>` | Switch to an existing user. |
| `users` | List all users, marking the one you're logged in as. |
| `addfeed <name> <url>` | Add an RSS feed and follow it. \* |
| `feeds` | List every feed and the user who added it. |
| `follow <url>` | Follow a feed that's already been added. \* |
| `following` | List the feeds you follow. \* |
| `unfollow <url>` | Stop following a feed. \* |
| `browse [limit]` | Show the newest posts from feeds you follow. Shows 2 if you leave out the limit. \* |
| `agg <interval>` | Fetch feeds in a loop, one per interval. The interval is a Go duration such as `30s`, `1m`, or `1h`. Runs until you stop it. |
| `reset` | Delete **all** users, along with their feeds, follows, and posts. |

\* Requires a logged-in user (see `register` and `login`).

Choose an `agg` interval that's polite to the sites you're fetching from. A minute or more is a reasonable default.

## Development

Run from source without installing:

```sh
go run . <command> [args...]
```

The SQL queries are in [`sql/queries`](sql/queries), and [sqlc](https://sqlc.dev) generates the Go code for them in [`internal/database`](internal/database). After changing a query or the schema, regenerate it:

```sh
sqlc generate
```
