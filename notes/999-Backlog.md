# Backlog

## Purpose

While working, I often note things to do that must be deferred until later to stay focused and make progress. I'm putting them here to makes finding them easier.

## Job Status Add

### Core features that go to other use cases in the future

* Consider writing a `gormRepo` that uses `gorm`.
  * It's an ORM that seems to have key features I'd want, but needs more investigation.
* Think about how to maintain and apply DDL

### DONE: Decide on a reliable approach for BusinessDate communication

`BusinessDate` is stored as `time.Time` in memory because Go doesn't have a bare Date data type.

Assumptions:

* We track processes that run in Singapore and Kansas City. Singapore is UTC +0800 and Kansas City is UTC -0600 (ignore daylight time for now).
* From the business perspective, business date "01 Jun 2023" is "01 Jun 2023" regardless of time zone. (Singapore 01 Jun 2023 may be over before KC 01 Jun 2023 starts.)
* The code may run on servers/containers that are NOT set to UTC or the local time zone for a given job status source (may not be in Singapore or US Central time).

I want to ensure the business date value I receive from the job status source is stored in the database in a time zone agnostic way so I shouldn't need to worry about whether the source is in Singapore or KC and SLO performance analysis is easier.

DONE: Added a Date type that handles date-only data to/from JSON and used it for `BusinessDate`. Now JSON can send `2023-06-15` and it is decoded and encoded correctly.

### DONE: Add error responses

When the job status source sends invalid job status data to the HTTP API, I want to respond with an appropriate error so my code can log and handle the error and so the job status source can respond to the error (log, alert, handle, etc.).

* If data is invalid, return a "bad data error" (custom) to the caller. (HTTP 400)
* If the row already exists, return a "duplicate row error" (custom) to the caller. (HTTP 409)
* If the database action fails for another reason, return a "database error" (custom) to the caller. (HTTP 500)

Custom errors define a `struct` and attach an `Error()` function to it. They can be returned as an `error`, but then they need to be type converted to get to the data. Maybe write a generic custom error that includes extra data I want for logging and a type, then I don't have to try for several possible type conversions. Also check `errors.Is()` and `errors.As()` and the idea of wrapping and unwrapping errors.

### DONE: When testing, ensure WithArgs() checks argument order

Because I'm using a database library that requires met to write SQL statements and pass arguments manually, SQL argument order is critical to ensure good data.

I want my test database mocks to check argument order so I'm sure the data that will be sent to the database is correct (don't cross column values).

## Simplify tests if possible

The tests for `database/sql` and `gorm` repos are almost identical. The differences are:

* `gorm` wraps the query in a transaction, so `db-mock` setup must account for it.
* `gorm` wants `defer db.Close()`.

If I wrap a transaction around the `database/sql` queries, which is probably a good idea, I can make the test bodies identical except `gorm` calls `gormBeforeEach()` and `database/sql` calls `dbSqlPgxBeforeEach`.

Can I somehow set the function the test uses in code so I don't need to duplicate test code?

## General

### DONE: Consider simplifying errors

I have `DomainError`, `AppError`, and `RepoError` so far. They're all the same. I'm not sure they'll ever diverge.

I want to answer the question, "Can I use a single custom error type?" so I can simplify errors if possible.

I want to promote primitive errors and codes to the common package to reduce duplication and make them available to tests.

### Decide where server startup(s) go

`go-slo` will end up running several different processes -- an HTTP API, consumers for the message bus, etc.

I want to be able to build different executables and deploy them as independently as possible so I can reduce deployment time and risk and have finer grained control of scaling.

This topic needs some research.

### DONE: Change naming pattern

Currently, I have names like `dbSqlPgRepo`, `memoryRepo`, and `serveHttpControllers`. This pattern makes it harder to find similar adapters.

I want to name files like `repoDbSqlPg`, `repoMemory`, `ctrlServeHttp` so similar adapters are grouped together.

Assumption: Given that I'm grouping modules by primary data object they serve, I'm more likely to have more than one `repo` or `ctrl` than I am to have more than one `dbSqlPg`, `serveHttp`, etc.

If I allow ingestion by HTTP, gRPC, and a public queue, I may have three controllers to adapt inbound data from them to the standard DTO. (When you get there, look at how to avoid duplicating error handling and messaging.)

If I support more than one backend database, I may have different repos for each to deal with their idioms.

### Add comparison capability to Date

The `Date` type I created doesn't compare with `time.Time` without casting.

I want to be able to compare `Date` and `time.Time` values without casting so code is less cluttered with casts and is easier to read and maintain.

I don't know if I can make `time.Compare` work with a `Date`, but I can have `Date.Compare` and `Date.CompareToTime`, for example.

### DONE: Pointer receivers vs. value receivers

When attaching methods to a `struct`, the receiver can be a pointer or value. Pointer receivers allow changing the instance referenced. Value receivers work on a copy.

```golang
type MyType struct{
  s string
}

// pointer receiver
func (mt *MyType) SetSPtr(str string) {
  mt.s = str
}

// value receiver
func (mt MyType) SetSVal(str string) MyType {
  mt.s = str
  return mt
}

func main() {
 v := MyType{s: "hello"}
 v.SetSVal("world1")
 fmt.Printf("%+v\n", v)

 v2 := v.SetSVal("world2a")
 fmt.Printf("%+v || %+v\n", v, v2)

 v = v.SetSVal("world2b")
 fmt.Printf("%+v\n", v)

 v.SetSPtr("world3")
 fmt.Printf("%+v\n", v)

 (&v).SetSPtr("world4")
 fmt.Printf("%+v\n", v)
}

/*** OUTPUT
{s:hello}
{s:hello} || {s:world2a}
{s:world2b}
{s:world3}
{s:world4}
***/
```

The compiler does some automatic conversions between pointers and values that work in some cases, but not others (interfaces). Common wisdom is to pick one and use it consistently for a given `struct` type, especially if an interface is in play. Many people lean toward pointers to have a common pattern everywhere.

My general preference is to avoid mutation, so I favor a strategy like the "world2a" and "world2b" examples above (explicitly overwrite). Unexpected mutation is a risk.

Golang does automatic conversions to/from pointers, so I might not know a function is mutating a value unless the function makes it explicit. I expect a setter to mutate and a getter not to mutate. Other functions may be unclear unless I explicitly passes a pointer (`someFunc(&mutateMe)` or `(&mutateMe).someFunc()`). Maybe an easy risk reduction is documentation comments and a consistent statement about whether a function mutates or not.

If the `struct` includes a synchronizing field (`sync.Mutex` or similar), methods must use a pointer receiver because copying the mutex breaks it. For `map`, `func`, `chan`, and slice, value receivers seem to be preferred. These type are the 1x - 3x the size of a pointer, so passing a pointer gets little gain and adds indirection. Small types (`int`, `rune`, etc.) fall into the same space as do small `structs`.

I also read a well reasoned article arguing against passing pointers seeking to avoid the cost of data copying. In modern CPU architectures, parameters passed by value are likely to be in L1/L2/L3 cache. Pointer parameters are more likely to require falling out to RAM, which is slower. The receiver is effectively another parameter to the function so behaves the same. Either decision seems like an optimization choice (value -> in cache -> better performance vs. pointer -> avoid copy -> better performance). There's a tradeoff and the correct answer is, "It depends."

I want to check all interfaces and methods that take a receiver and make them consistent so behavior is predictable, making decisions based on sound understanding of what they need to do.

I want to ensure all methods and interfaces have documentation comments that include a statement about whether a function mutates the receiver or not.
