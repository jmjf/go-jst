# Setup

I want to build an application using golang, Postgres, and Kafka topics. I have the design sketched and will write it up in `001-Design.md`.

I set up a dev container that uses `docker-compose.postgres.yml` and `docker-compose.golang.yml`. That makes it easy to try different database and message bus options by changing the compose files `devcontainer.json` uses.

First, I want to get my dev container setup and be able to connect to pg from a simple program. I started with `pq`, but noticed it's in maintenance mode and recommended [`pgx`](https://github.com/jackc/pgx), so switched to `pgx`.

I chose to use `pgx` with golang's `database/sql`. While I don't plan to use other databases, I'm going with the more flexible approach because it's likely more portable and valuable.

I used `adminer` to create a table and insert a row into it.

```sql
CREATE TABLE "public"."JobStatus" (
    "JobId" integer NOT NULL,
    "StatusCode" character varying(10) NOT NULL,
    "StatusTimestamp" timestamptz NOT NULL,
    "BusinessDate" date NOT NULL
) WITH (oids = false);

INSERT INTO "JobStatus" ("JobId", "StatusCode", "StatusTimestamp", "BusinessDate") VALUES
(1, 'SUCCEED', '2023-06-16 00:18:33.324286+00', '2023-06-15');
INSERT INTO "JobStatus" ("JobId", "StatusCode", "StatusTimestamp", "BusinessDate") VALUES
(2, 'SUCCEED', '2023-06-16 01:18:33.324286+00', '2023-06-16');
```

I set up a simple `go.mod` and ran `go get github.com/jackc/pgx/v5` to get `pgx`. Then I wrote `testdb.go` based on a couple of examples. A few key points I noted.

* `pgUrl` follows a standard Postgres connection URL pattern. I need to investigate security implications or alternatives. (Even in HTTPS, URLs are plain text.)
* `"pgx"` in `sql.Open()` tells `database/sql` to use `pgx`.
* `database/sql` can query single rows with `db.QueryRow().Scan()` -- not used here, but noted for single row results.
* `db.Query` returns a `rows` pointer that has the result object. It's treated like a cursor. I want to know if it uses a cursor under the covers or is getting all results into memory.
* To read `rows`, first we need to call `rows.Next()`, then we can `rows.Scan()` the row.
* The list of arguments passed to `rows.Scan()` must match the result set of the query or the program will panic.
* When pg returns a Date, it becomes a `time.Time`. Use the `.Format()` method to make it look like a date.
* Time formats are described using a reference date, which is 2006-01-02 15:04:05 -0700. So, `"2006-01-02"` is YYYY-MM-DD, etc.
* `rows` must be closed with `rows.Close()`, so `defer` it after querying.
* `rows.Columns()` returns an array of column name strings and an error.

## Output

```bash
dev@8cf952d4217b:/workspace$ go run testdb.go
Connect to  postgres://postgres:postgres@192.168.0.101:9432/gojst
Query
[JobId StatusCode StatusTimestamp BusinessDate]
1 | SUCCEED | 2023-06-16 00:18:33.32 +0000 | 2023-06-15
2 | SUCCEED | 2023-06-16 01:18:33.32 +0000 | 2023-06-16
Successful connection
```

## Things to investigate

* Is the URL approach safe? (Does it protect the username and password under the covers?)
* Does `db.Query()` act like a cursor or actually use a cursor?
* Is there a way to use `Scan()` with an incomplete argument list?

**COMMIT: CHORE: setup and test pg connection**
