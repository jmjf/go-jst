package jobStatus

import (
	"database/sql"
	"time"
)

type dbSqlPgRepo struct {
	db                        *sql.DB
	sqlInsert                 string
	sqlSelect                 string
	sqlWhereJobId             string
	sqlWhereJobIdBusinessDate string
}

// NewDbSqlRepo creates a new dbSqlRepo object using the passed database handle.
// Passing the handle lets it be setup during application startup and shared with other repos.
func NewDbSqlPgRepo(db *sql.DB) JobStatusRepo {
	return &dbSqlPgRepo{
		db: db,

		// The order of columns in the following statements is significant.
		// The insert operation uses a set of values from dbToDomain, which assumes a specific order of columns.
		// The select operation scans data from the result set assuming a specific order of columns.
		// ALWAYS use the same order in all statements!

		sqlInsert: `
			INSERT INTO "JobStatus" ("ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId")
			VALUES($1, $2, $3, $4, $5, $6, $7)
		`,
		sqlSelect:                 `SELECT "ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId" FROM "JobStatus"`,
		sqlWhereJobId:             `WHERE "JobId" = $1`,
		sqlWhereJobIdBusinessDate: `WHERE "JobId" = $1 AND "BusinessDate" = $2`,
	}
}

// add inserts a JobStatus into the database.
func (repo dbSqlPgRepo) add(jobStatus JobStatus) error {
	// we only care that it succeeds, not looking for a return, so use Exec()
	_, err := repo.db.Exec(repo.sqlInsert, domainToDb(jobStatus)...)
	// jobStatus.ApplicationId, jobStatus.JobId,
	// jobStatus.JobStatusCode, jobStatus.JobStatusTimestamp,
	// jobStatus.BusinessDate, jobStatus.RunId, jobStatus.HostId)
	if err != nil {
		return err
	}

	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
func (repo dbSqlPgRepo) GetByJobId(jobId JobIdType) ([]JobStatus, error) {
	rows, err := repo.db.Query(repo.sqlSelect+repo.sqlWhereJobId, jobId)
	if err != nil {
		return []JobStatus{}, err
	}
	defer rows.Close()

	return rowsToSlice(rows)
}

// GetByJobIdBusinessDate retrieves JobStatus structs for a specific job id and business date.
func (repo dbSqlPgRepo) GetByJobIdBusinessDate(jobId JobIdType, busDt time.Time) ([]JobStatus, error) {
	rows, err := repo.db.Query(repo.sqlSelect+repo.sqlWhereJobIdBusinessDate, jobId, busDt)
	if err != nil {
		return []JobStatus{}, err
	}
	defer rows.Close()

	return rowsToSlice(rows)
}

// rowsToSlice converts the database job status data in rows to a usable slice of JobStatus structs.
//
// If dbToDomain() fails to convert any row in the result set, it returns an empty slice and an error.
func rowsToSlice(rows *sql.Rows) ([]JobStatus, error) {
	var result []JobStatus

	for rows.Next() {

		jobStatus, err := dbToDomain(rows)
		if err != nil {
			return []JobStatus{}, err
		}

		result = append(result, jobStatus)
	}
	return result, nil
}

// dbToDomain converts database job status data to a JobStatus struct by scanning rows for values and building JobStatus.
func dbToDomain(rows *sql.Rows) (JobStatus, error) {
	var (
		appId string
		jobId JobIdType
		jobSt JobStatusCodeType
		jobTs time.Time
		busDt time.Time
		runId string
		hstId string
	)

	err := rows.Scan(&appId, &jobId, &jobSt, &jobTs, &busDt, &runId, &hstId)
	if err != nil {
		return JobStatus{}, err
	}

	return JobStatus{
		ApplicationId:      appId,
		JobId:              jobId,
		JobStatusCode:      jobSt,
		JobStatusTimestamp: jobTs,
		BusinessDate:       busDt,
		RunId:              runId,
		HostId:             hstId,
	}, nil
}

// domainToDb converts a JobStatus into an array of values to insert.
//
// SQL statements that specify values must use the expected order.
//
// Expected order:
//
//	ApplicationId, JobId, JobStatusCode, BusinessDate, RunId, HostId
func domainToDb(jobStatus JobStatus) []any {
	return []any{
		jobStatus.ApplicationId,
		jobStatus.JobId,
		jobStatus.JobStatusCode,
		jobStatus.JobStatusTimestamp,
		jobStatus.BusinessDate,
		jobStatus.RunId,
		jobStatus.HostId,
	}
}
