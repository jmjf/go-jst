# Phase 1 -- Job Status

## Workspace

I did a [side experiment on workspaces](https://github.com/jmjf/go-workspace-experiment) to understand how they work. I plan to have several executables in this bounded context and want to share common code. So, I need to define my modules.

In the same experiment, I sorted out a basic approach for using clean/DDD/hexagonal/whatever architecture, which I'll be using here.

Here's the logical entities I know about:

* `jobStatus` -- includes references to `application` and `job`
* `application`
* `job`
* `slo` -- includes references to `application` and `sloToJob`
* `sloToJob` -- includes references to `slo` and `job`
* `sloPerformance` -- summary of meet/miss data for each SLO period

To resolve the circular reference between `slo` and `sloToJob`, I'll probably make `slo` an aggregate that includes many `sloToJob` in memory. In the database, it's two different tables. I'll also `job` to have `sloToJob`, so I think I'll have something like this:

```golang
// naming pattern TBD, but this way is good for now to draw distinctions

type SloEntity struct {}
type JobEntity struct {}
type SloJobEntity struct {} // carries ids, not jobs or SLOs
type ApplicationEntity struct {}
type JobStatusEntity struct {}
type SloPerformanceEntity struct {}

type SloAggregate struct{
   Slo SloEntity
   SloJobs []SloJobEntity // we need the relationship entity because it has expected start/end times
   // SLO performance could get large so deal with it in slices that have the range of time of interest
}

type JobAggregate struct {
   Job JobEntity
   JobSlo []SloJobEntity // the relationship can go either way
}

type ApplicationAggregate struct {
   Application ApplicationEntity
   ApplicationSlos []SloAggregate // simple relationship with no extra attributes
}
```

I reserve the right to change my mind as I move forward and better understand how to manage the data.

For phase 1, I'm concerned with `JobStatus`. Data from HTTP includes ids for `job` and `application`. I may build an aggregate that includes their details later in phase 1.

So, let's expect we'll have modules for `JobStatus`, `Slo`, `Job` and `Application` at a minimum. For phase 1 I'll focus on `JobStatus` and build the rest out when I'm ready for them.

We can set up the workspace.

```bash
# Cleanup from earlier testing
rm go.mod
rm go.sum
rm testdb.go # it's in commit history when I want to see it

# Create the workspace
go work init
mkdir jobStatus && cd jobStatus && go mod init jobStatus
cd ..
```

I'll sort out how to manage the pg connection later. I may try to build a shared connection object. TBD.

## Job status basics

`domain.go` defines the domain types, constants, and common interfaces

* Basic types
* Constants -- all constants for the module begin with `JobStatus_` and end with an all-caps name (ex: `JobStatus_START`)
* Structs -- `JobStatus` and `JobStatusDto`; latter has JSON tags and uses short field names
* Interfaces -- `JobStatusRepo` and `JobStatusUC`

Build a repo using `database/sql` in `dbSqlRepo.go`. Build a repo using a slice in memory in `memoryRepo.go`. When I start the application, I can create the repo I prefer. Use cases will use the interface, so won't care which implementation I'm using.

For now, I'm leaving `go.work` out of `.gitignore`. Also, I'm changing how I tag commits in markdown files to avoid complaints about full-line emphasis.

**COMMIT:** FEAT: add repos for database/sql and memory; define key data structures; establish workspace

## Summary so far

What's going on and why am I breaking things up this way?

Clean/DDD/etc. architecture models recommend separating implementation details from business logic. That leads to patterns like:

* **Entities**, **aggregates** and similar data structures that include core business logic for managing the data they hold.
  * I haven't build the intelligence on `JobStatus` yet, but it's coming.
  * Aggregates may be composed of entities. They may carry domain event information.
* **Use cases**, services or similar that perform business processes on the entities/aggregates to meet business requirements.

These two types of objects make up the business domain space. Arguably, a non-technical user could read their names and possibly their code and understand them. The domain space may include **domain events** used to notify the system of actions. For example, in `go-slo`, we might have a domain event for "job status created" or "SLO missed" that can be published to subscribers to trigger actions.

* Repos, controllers and other **adapters** that manage the interface between use cases (process) and the outside world (database, HTTP, message bus, logging, etc.)
  * Logging happens in the adapters, not in use cases.
  * Adapters are responsible for translating data from raw, external formats to a form the business domain code recognizes.
* **Infrastructure** represents the outside world and includes database drivers, HTTP servers, etc., that the adapters use.

Separating business logic from external infrastructure lets us change infrastructure by building new adapters. Today we're developing on SQLite. Tomorrow we decide to switch to Postgres or MySQL as we approach release. Next year, our SaaS platform is the new hot thing so we're running on a cloud database or need to support different databases due to different customers' demands. The business logic is decoupled from the infrastructure, allowing them to change independently.

This separation is made possible by defined interfaces. So, look at the `JobStatusRepo` in `domain.go`. We can provide anything that satisfies that interface to our use cases and they won't care about the physical implementation details. Look at `dbSqlRepo.go` and `memoryRepo.go`. They're completely different ways to store and manage data, but I can give either to a use case as a `JobStatusRepo` and the use case will work with no changes. And if I decided to replace `database/sql` with direct `pgx` or `gorm` or MySQL or MongoDB, as long as the adapter is a `JobStatusRepo`, the domain code doesn't care.

Return patterns are part of the interface. Adapters must have a consistent definition of an error. If "not found" returns an error in `dbSqlRepo`, it needs to return an error in other repos or `dbSqlRepo` needs to convert the error to an empty result on return.

## Create tables and write a program to perform simple tests

I noticed I forgot to include the job status timestamp, which tells us when the job reached the status. I've added that in the `dbSqlRepo`. It isn't an issue in `memoryRepo` because it does no data mapping (yet).

The `JobStatus` table has the following DDL. Postgres defines the index as a primary key, which means the combination of columns must be unique. For our testing purposes, that's fine, but I'll need to think about indexing later. (Premature indexing is a root of many evils and much wasted space and server CPU and I/O cycles.)

```sql
CREATE TABLE "public"."JobStatus" (
    "ApplicationId" character varying(200) NOT NULL,
    "JobId" character varying(200) NOT NULL,
    "JobStatusCode" character varying(10) NOT NULL,
    "JobStatusTimestamp" timestamptz NOT NULL,
    "BusinessDate" date NOT NULL,
    "RunId" character varying(50),
    "HostId" character varying(150),
    CONSTRAINT "JobStatus_Primary" PRIMARY KEY ("JobId", "BusinessDate", "JobStatusTimestamp", "ApplicationId")
) WITH (oids = false);
```

I'm testing directly with the repos, so for this test, I'm making the `JobStatusRepo`'s `add` method exported by changing it to `Add`.

I built `./cmd/testRepo/testRepo.go` to test both repos. Both repos using the same test function, which accepts a repo to test. Think of the test function as a use case using the repo. It doesn't care which repo it uses.

Important things I learned:

* MySQL uses `?` placeholders in parameterized queries, but Postgres uses `$1`, `$2`, etc. So I renamed `dbSqlRepo` to `dbSqlPgRepo` because the query strings in it are Postgres-specific.
* Either `database/sql` or `pgx` is converting table and column names to all lower case unless they're in double-quotes. Postgres is set to run case sensitive, so I added double-quotes in all the query strings.

After some tweaking, both `testRepo` tests `dbSqlPgRepo` and `memoryRepo` with no problems. When testing, ensure the `JobStatus` table is empty to avoid duplicate key errors.

**COMMIT:** TEST: run tests for both repos and confirm implementation agnosticism

## Use cases

Job status is pretty simple. We can add them and read them. There's no update and any deletes will happen as part of a scheduled purge process.

In `domain.go` we have a data transfer object (DTO), `JobStatusDto`. It carries data into the use cases from HTTP or other sources.

## Add use case

The `Add` use case stores a new job status row in the database so the data is available for job status history and SLO performance calculation (if reporting from the database). At a basic level, it needs to:

* Ensure times are good (need for JobTs/BusDt validation to work)
* Check the data in the DTO to be sure it's usable.
* Create a `JobStatus`
* Ensure the row to be inserted isn't a duplicate.
* Insert the row in the database.

### Ensure times are good

`JobStatusTimestamp` and `BusinessDate` may arrive with too much precision. Also, life is easier if all times in the database are normalized to UTC. Reports and a future UI can convert to/from a specified time zone if needed.

* `JobStatusTimestamp`
  * Set to `JobStatusTimestamp.Truncate(time.Second).UTC()` -- truncates to 1 second accuracy and converts to UTC
* `BusinessDate`
  * `yr, mo, dy := BusinessDate.Date()` then set to `time.Date(yr, mo, dy, 0, 0, 0, 0, time.UTC)`
  * `BusinessDate` is an absolute date not affected by time zone shifts, so we can normalize to UTC this way.

Written as `normalizeTimes`, a method on the DTO in `dataObjects.go`. Because these rules are business data rules, I'm treating the DTO definition as a domain object. I may change my mind on how to handle this later, but it makes sense to me for now.

### Check data in DTO

What checks do we need to validate the DTO?

* `ApplicationId`
  * Not empty
  * Not too long
  * Is a known id (skip for now because no data)
* `JobId`
  * Not empty
  * Not too long
  * Is a known id (skip for now because no data)
  * Assumes we'll have data for all jobs we care about and won't store data for jobs we don't care about
* `JobStatusCode`
  * Is found in the array of valid statuses
  *Added `ValidJobStatusCodes`; decide if export is needed later.
* `JobStatusTimestamp`
  * Not empty
  * Not in the future
* `BusinessDate`
  * Not empty
  * Not in the future
  * Is <= the date part of `JobStatusTimestamp` (need to be sure this doesn't cause problems with Asia)
* `RunId` is not too long
* `HostId` is not too long

FUTURE: If data is invalid, return a "bad data error" (custom) to the caller for logging and return to client. (HTTP 400)

Written as `isUsable`, a method on the DTO in `dataObjects.go`. Same reasoning as 'normalizeTimes`.

### Create JobStatus

The decisions above lead to some changes.

"Can I get a good `JobStatus`?" is a question for the domain, not the use case. The use case attempts to create a `JobStatus` and either gets a `JobStatus` it can use or gets an error.

So, the use case is now:

* Try to create a `JobStatus` with the DTO.
  * On error, return error
* Try to add the `JobStatus` to the database.
  * On error, return error
  * ASSUMPTION: attempting to add a duplicate row will return an error, so no need for a separate duplicate check
* Return the `JobStatus`.

Because I've moved data checks onto the DTO, I could call the DTO functions in the use case and keep the validation there. That might make sense, but I think it makes more sense from a business reasoning perspective to say, "When I try to create a `JobStatus`, I ensure the data I'm trying to use is valid." The alternative is scattering the validation in all the places we might create a `JobStatus`. For example, the repo needs to call `newJobStatus` with the raw data passed as a DTO to get a `JobStatus` to return rather than just trusting the data in the database.

### Definition of a duplicate row

We don't want to insert a row in the database where `ApplicationId`, `JobId`, `JobStatusCode`, `BusinessDate` and `JobStatusTimestamp` are the same because that can cause problems when calculating SLO performance. I could do this with a primary key constraint on the table in the database. That's probably less overhead that querying to check for a duplicate.

FUTURE: If the row already exists, return a "duplicate row error" (custom) to the caller for logging and return to client. (HTTP 409)

### Add JobStatus

Call the repo's add method.

FUTURE: Need a "database error" (custom) that the repo can return so the caller can log correctly. (HTTP 500)

**COMMIT:** FEAT: add the Add use case and reorganize code based on where it led thinking

## Testing

Investigate how to write tests and options.

I can test the domain and leave the use cases untested, but that seems questionable if the UCs end up carrying business process logic.

How might I mock or intercept calls to the database. I'd really like the repo code to be exercised too so I can confirm it returns errors correctly. An ORM could standardize errors to some extent, but I'm not sure I'm ready to go that route yet.

## Next

* Define and write use cases with unit tests.
  * How do I deal with the repo and intercepting calls or mocking?
* Change repos to use `newJobStatus` when getting data.
* Write an HTTP server for job status ingestion.
  * Write controllers and figure out controller structure
  * Add logging middleware to log requests, responses, and response times (shared library module).
* Investigate structured logging options and how to carry logging info for errors. Build custom errors and use them.
  * BadDataError
  * DatabaseError
  * DuplicateRowError (is a DatabaseError)
* Consider writing a `gormRepo` that uses `gorm`.
  * It's an ORM that seems to have key features I'd want, but needs more investigation.
* Think about how to maintain and apply DDL

## Notes

Custom errors define a `struct` and attach an `Error()` function to it. They can be returned as an `error`, but then they need to be type converted to get to the data. Maybe write a generic custom error that includes extra data I want and a type, then I don't have to try for several possible type conversions. Also check `errors.Is()` and `errors.As()` and the idea of wrapping and unwrapping errors.
