# Starting plan

This plan is the starting plan because I may add more features later.

## Problem

Foxfire Finance's systems run many batch jobs for business operational and reporting purposes.

A job may include many steps and the output of one job may feed another.

For example, the job that calculates and charges overdrafts at the end of the day produces data that feeds decision support systems. Other jobs use data from the decision support systems to summarize activity so Foxfire can understand overdraft activity and the economic effects it might indicate.

Another job runs every 30 minutes during from 6:00 a.m. to 6:00 p.m. to summarize checking and savings transaction activity for management dashboards. The 6:00 a.m. run summarizes data for each 30 minute period overnight.

To manage expectations, each job has a service level objective (SLO) to deliver data by a certain time.

For example, the overdraft job should produce data to load the decision support system by 3:00 a.m every calendar day. The job that loads the data into the decision support system is due by 5:00 a.m each business day. The report runs on back office business days by 8:00 a.m.

The checking and savings transaction summary and load to dashboard system updates by 10 minutes after and 40 minutes after each hour.

Foxfire wants an application to track job SLO performance to identify:

* Jobs that are late so they can understand consequences and act accordingly
* Jobs that are slowing over time which may cause SLO misses in the future
* Jobs that are routinely early or late to consider changing SLOs or improving infrastructure running the jobs to meet SLOs.

## Data

**SLO**

* An SLO id
* An SLO name
* The id of the business application the SLO is associated with (overdrafts, deposits, etc.)
* SLO calendar -- defines the dates on which it runs.
  * Most SLOs run on a standard calendar that takes into account holidays.
  * Some SLOs may run on one of several alternate calendars (include weekends, include holidays, etc.).
* Relationships to one or more jobs upon which the SLO depends.

**SLO To Job Relationship**

* An SLO id
* A Job id
* An expected start time
* An expected end time

An SLO is complete only when all jobs related to it are complete. More than one SLO may depend on the same job, especially in chained processes like the overdraft process, load to decision support, report chain.

Foxfire's job schedulers or the jobs themselves send job status data to the SLO tracking system.

**Job Status**

* Application id
* Job id
* Status code (start, succeed, fail)
* Status time (date + time)
* Business date -- if the overdraft job for Monday starts at 12:30 a.m. Tuesday, the job runs on Tuesday for business date Monday.
* Run id -- may change if a job fails and is restarted
* Host id -- where the job runs

The system uses SLO definitions and job status data to build an SLO performance table for reporting.

**SLO Performance**

* SLO id
* Business date
* Expected start time (earliest expected start time for all jobs)
* Expected end time (latest expected end time for all jobs)
* Actual start time (earliest start time or null no job is not started)
* Actual end time (latest end time or null if any jobs are not complete)

## Build plan

### Phase 1 -- post job status to database

* Job status data will be sent to an HTTP API that will store the status in the database
* All jobs start and end within one job frequency unit.
  * If a job runs daily, the job will finish before the next day's job runs
* SLO performance reporting will be done manually.

### Phase 2 -- calculate SLO performance

* The job status API will make an HTTP POST call to an SLA performance calculator service.
* The SLO performance calculator will update the SLO performance table.
* SLO performance reporting will happen by querying the table (or by a performance view that derives columns based on the SLO Performance table).

### Phase 3 -- replace SLO performance HTTP call

* The job status API will publish the job status on a message bus
* The SLO performance service will read data from the message bus

### Phase 4 -- decouple job status and SLO performance updates

* The HTTP API will perform basic data quality checks and publish data to a message bus.
* A job status update service will read data from the message bus and update the job status table.
* The SLO performance service will read data from the message bus an update SLO performance.

## Possible future enhancements

* Allow SLO start/end times to be offset from the business date to account for delayed or long running processes.
* Add to SLOs a "may be this late without affecting other SLOs" time and notify operations staff if that window is breached.
  * Notify can be a log message for demo purposes
* Security -- clients must pass a token that identifies them; client identities are authorized to update status for certain applications only.
  * I may start with a "fake" token and move to OAuth2 later.
  * Put AuthZ in a shared service.
* Adapter services for data from other sources (Kafka, other messaging systems, files, logs, etc.)
* More TBD (will be PlanB, etc.)

## Nonfunctional requirements

* Design with clean architecture in mind
  * Domain objects that know how to create, validate, etc., themselves
  * Use cases that marshall domain objects and perform processes
  * Database access through repos
  * HTTP/message preprocessing in controllers
  * Consider separating controllers from receivers (i.e., handleFunc() will call the controller, not include the controller code) -- TBD
* Identify common components and reuse them across services
* Must have unit tests to confirm the core business logic works as expected.
* Must have single-service integration tests to ensure the service fulfills any contracts.
* Performance requirements TBD
* Must be able to run more than one instance of each service
* Must have structured logging to support application behavior analysis and observation
  * Start with console logging
  * Later to a log aggregator/searcher
* A few things I'm forgetting

## Other notes

* Until proven otherwise, this application represents a single bounded context.
  * Services share read access to tables. (Example: SLO performance and notification are likely to need read access to SLO data.)
  * BUT, only one service has write responsibilities to a table.
* Golang
* Postgres database using `database/sql` and `pgx` (like `testdb.go`)
* Use native Golang http for now (no `gin`, etc.)
  * I want to see how it works before considering detail-hiders/simplifiers.
* Kafka messaging (package TBD)
* Plain HTTP until adding authN/authZ
* Other decisions will be made and documented as I work.

**COMMIT: DOCS: describe the product; plan starting features**
