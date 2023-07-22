package repo

import (
	"database/sql"
	"fmt"
	"time"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

type repoDB struct {
	DSN                string
	DB                 *sql.DB
	sqlInsert          string
	sqlSelect          string
	sqlWhereJobId      string
	sqlWhereJobIdBusDt string
}

// NewRepoDb creates a new database/ORM specific object using the passed DSN.
// Passing the handle lets it be setup during application startup and shared with other repos.
func NewRepoDB(DSN string) *repoDB {
	return &repoDB{
		DSN: DSN,

		// The order of columns in the following statements is significant.
		// The insert operation uses a set of values from dbToDomain, which assumes a specific order of columns.
		// The select operation scans data from the result set assuming a specific order of columns.
		// ALWAYS use the same order in all statements!

		sqlInsert: `
			INSERT INTO "JobStatus" ("ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId")
			VALUES($1, $2, $3, $4, $5, $6, $7)
		`,
		sqlSelect: `SELECT "ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId" FROM "JobStatus"`,
	}
}

// Open connects to the database described by the dsn set on the repo.
//
// Mutates receiver: yes (sets repo.DB)
func (repo *repoDB) Open() error {
	if repo.DSN == "" {
		return internal.NewLoggableError(internal.ErrRepoNoDsn, internal.ErrcdRepoNoDsn, nil)
	}

	db, err := sql.Open("pgx", repo.DSN)
	if err != nil {
		return internal.NewLoggableError(err, internal.ErrcdRepoConnException, nil)
	}
	repo.DB = db
	return nil
}

// Close closes the repo's database connection
//
// Mutates receiver: no
func (repo *repoDB) Close() error {
	if repo.DB != nil {
		return repo.DB.Close()
	}
	return nil
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: no
func (repo *repoDB) Add(jobStatus jobStatus.JobStatus) error {
	// we only care that it succeeds, not looking for a return, so use Exec()
	_, err := repo.DB.Exec(repo.sqlInsert, domainToDb(jobStatus)...)
	if err != nil {
		code := internal.PgErrToCommon(err)
		return internal.NewLoggableError(err, code, jobStatus)
	}
	return nil
}

// GetByQuery retrieves JobStatus structs for query criteria expressed in the dto.
// Currently only supports JobId and BusDt
// TODO: Handle other values
//
// Mutates receiver: no
func (repo *repoDB) GetByQuery(q map[string]string) ([]jobStatus.JobStatus, error) {
	where, vals, err := queryToWhere(q)
	if err != nil {
		internal.NewLoggableError(err, internal.ErrcdRepoInvalidQuery, map[string]any{"err": err, "query": q})
	}

	rows, err := repo.DB.Query(repo.sqlSelect+where, vals...)
	if err != nil {
		code := internal.PgErrToCommon(err)
		return nil, internal.NewLoggableError(err, code, map[string]any{"where": where, "vals": vals})
	}
	defer rows.Close()

	data, err := rowsToDomain(rows)
	if err != nil {
		return nil, internal.WrapError(err)
	}
	if len(data) == 0 {
		return nil, internal.NewLoggableError(internal.ErrRepoNotFound, internal.ErrcdRepoNotFound, map[string]any{"where": where, "vals": vals})
	}
	return data, nil
}

// queryToWhere builds a where clause and values from the query map, which has all values as strings.
func queryToWhere(q map[string]string) (string, []any, error) {
	var vals []any
	fieldNum := 1
	where := " WHERE " // leading space because it will be added to SELECT

	for nm, val := range q {
		dbVal, err := fieldToDb(nm, val)
		if err != nil {
			return "", nil, err
		}
		if fieldNum > 1 {
			where += " AND "
		}
		where += fmt.Sprintf(`"%s" = $%d`, nm, fieldNum)
		vals = append(vals, dbVal)
		fieldNum++
	}
	return where, vals, nil
}

// fieldToDb converts string values from the query map to database values
// we can use to run queries.
func fieldToDb(nm string, val string) (any, error) {
	// explicitly convert fields that are not string
	switch nm {
	case "JobStatusTimestamp": // time.Time
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return t.UTC(), nil
	case "BusinessDate": // internal.Date
		dt, err := internal.NewDate(val)
		if err != nil {
			return nil, err
		}
		return dt.AsTime(), nil
	default: // all the rest are strings
		return val, nil
	}
}

// rowsToDomain converts a slice of database job status data to a slice of domain data by calling dbToDomain() for each item.
// If dbToDomain() fails to convert any row in the result set, it returns an empty slice and an error.
func rowsToDomain(rows *sql.Rows) ([]jobStatus.JobStatus, error) {
	var result []jobStatus.JobStatus

	for rows.Next() {

		jobStatus, err := dbToDomain(rows)
		if err != nil {
			return nil, internal.WrapError(err)
		}

		result = append(result, jobStatus)
	}
	return result, nil
}

// dbToDomain converts database job status data to a JobStatus struct by scanning rows for values and building JobStatus.
func dbToDomain(rows *sql.Rows) (jobStatus.JobStatus, error) {
	var (
		appId string
		jobId jobStatus.JobIdType
		jobSt jobStatus.JobStatusCodeType
		jobTs time.Time
		busDt time.Time // database/sql will Scan to time.Time, not internal.Date
		runId string
		hstId string
	)

	err := rows.Scan(&appId, &jobId, &jobSt, &jobTs, &busDt, &runId, &hstId)
	if err != nil {
		return jobStatus.JobStatus{}, internal.NewLoggableError(err, internal.ErrcdRepoScan, rows)
	}

	return jobStatus.NewJobStatus(dtoType.JobStatusDto{
		AppId: appId,
		JobId: string(jobId),
		JobSt: string(jobSt),
		JobTs: jobTs,
		BusDt: internal.NewDateFromTime(busDt),
		RunId: runId,
		HstId: hstId,
	})
}

// domainToDb converts a JobStatus into an array of values to insert.
// SQL statements that specify values must use the expected order.
//
// Expected order: ApplicationId, JobId, JobStatusCode, BusinessDate, RunId, HostId
func domainToDb(jobStatus jobStatus.JobStatus) []any {
	return []any{
		jobStatus.ApplicationId,
		jobStatus.JobId,
		jobStatus.JobStatusCode,
		jobStatus.JobStatusTimestamp,
		jobStatus.BusinessDate.AsTime(),
		jobStatus.RunId,
		jobStatus.HostId,
	}
}
