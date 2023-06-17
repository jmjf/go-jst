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

**COMMIT: CHORE: setup and test pg connection**

## Networking change

I was using the host machine's IP address as `host`. I didn't like that, so I ran `docker container inspect` with a partial container id (easier than typing the name). That got me output that included:

```json
    "Networks": {
        "go-jst_devcontainer_default": {
            "IPAMConfig": null,
            "Links": null,
            "Aliases": [
                "1a91b0cacdc6",
                "db"
            ],
            "NetworkID": "bd1610120990a5c695434c87c2863b68fd77e31eb8a4f6ce498fdcd30895430e",
            "EndpointID": "9913853b9c065abe22f9d1fc2a175fc7d578af1366a3313017fcba269bfb9997",
            "Gateway": "172.18.0.1",
            "IPAddress": "172.18.0.2",
            "IPPrefixLen": 16,
            "IPv6Gateway": "",
            "GlobalIPv6Address": "",
            "GlobalIPv6PrefixLen": 0,
            "MacAddress": "02:42:ac:12:00:02",
            "DriverOpts": null
        }
    }
```

I confirmed `ping 172.18.0.2` (the container's IP address) and `ping db` both worked and were both pinging the same IP.

I changed the `host` and `port` based on that test. I changed the `port` because this change routes inside the Docker network, which means we get the inside port (5432), not the outside port (9432) from `docker-compose.postgres.yml`.

```yaml
    ports:
      - 9432:5432
```

**COMMIT: FIX: use the container network instead of the host network to connect to Postgres**

## Things to investigate

* Is the URL approach safe? (Does it protect the username and password under the covers?)
* Does `db.Query()` act like a cursor or actually use a cursor?
* Is there a way to use `Scan()` with an incomplete argument list?
