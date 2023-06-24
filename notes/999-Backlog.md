# Backlog

## Purpose

While working, I often note things to do that must be deferred until later to stay focused and make progress. I'm putting them here to makes finding them easier.

## Job Status Add

### Next

* Change repos to use `newJobStatus` in `dbToDomain()`.
* Write an HTTP server for job status ingestion.
  * Write controllers and figure out controller structure
* Add Logging
  * Build middleware to log requests, responses, and response times (shared library module).
  * Decide how to assign ids to requests.
  * Investigate structured logging options and how to carry logging info for errors. Build custom errors and use them.
    * BadDataError
    * DatabaseError
    * DuplicateRowError (is a DatabaseError)
* Consider writing a `gormRepo` that uses `gorm`.
  * It's an ORM that seems to have key features I'd want, but needs more investigation.
* Think about how to maintain and apply DDL

### Decide on a reliable approach for BusinessDate communication

`BusinessDate` is stored as `time.Time` in memory because Go doesn't have a bare Date data type.

Assumptions:

* We track processes that run in Singapore and Kansas City. Singapore is UTC +0800 and Kansas City is UTC -0600 (ignore daylight time for now).
* From the business perspective, business date "01 Jun 2023" is "01 Jun 2023" regardless of time zone. (Singapore 01 Jun 2023 may be over before KC 01 Jun 2023 starts.)
* The code may run on servers/containers that are NOT set to UTC or the local time zone for a given job status source (may not be in Singapore or US Central time).

I want to ensure the business date value I receive from the job status source is stored in the database in a time zone agnostic way so I shouldn't need to worry about whether the source is in Singapore or KC and SLO performance analysis is easier.

### Add error responses

When the job status source sends invalid job status data to the HTTP API, I want to respond with an appropriate error so my code can log and handle the error and so the job status source can respond to the error (log, alert, handle, etc.).

* If data is invalid, return a "bad data error" (custom) to the caller. (HTTP 400)
* If the row already exists, return a "duplicate row error" (custom) to the caller. (HTTP 409)
* If the database action fails for another reason, return a "database error" (custom) to the caller. (HTTP 500)

Custom errors define a `struct` and attach an `Error()` function to it. They can be returned as an `error`, but then they need to be type converted to get to the data. Maybe write a generic custom error that includes extra data I want for logging and a type, then I don't have to try for several possible type conversions. Also check `errors.Is()` and `errors.As()` and the idea of wrapping and unwrapping errors.

### When testing, ensure WithArgs() checks argument order

Because I'm using a database library that requires met to write SQL statements and pass arguments manually, SQL argument order is critical to ensure good data.

I want my test database mocks to check argument order so I'm sure the data that will be sent to the database is correct (don't cross column values).

## General

### Decide where server startup(s) go

`go-slo` will end up running several different processes -- an HTTP API, consumers for the message bus, etc.

I want to be able to build different executables and deploy them as independently as possible so I can reduce deployment time and risk and have finer grained control of scaling.

This topic needs some research.
