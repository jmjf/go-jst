package jobStatus

import (
	"common"
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
//
// Mutates receiver: no
func (repo dbSqlPgRepo) add(jobStatus JobStatus) error {
	// we only care that it succeeds, not looking for a return, so use Exec()
	_, err := repo.db.Exec(repo.sqlInsert, repo.domainToDb(jobStatus)...)
	if err != nil {
		code := common.PgErrToCommon(err)
		return common.NewCommonError(err, code, jobStatus)
	}
	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: no
func (repo dbSqlPgRepo) GetByJobId(jobId JobIdType) ([]JobStatus, error) {
	rows, err := repo.db.Query(repo.sqlSelect+repo.sqlWhereJobId, jobId)
	if err != nil {
		code := common.PgErrToCommon(err)
		return []JobStatus{}, common.NewCommonError(err, code, map[string]any{"jobId": jobId})
	}
	defer rows.Close()

	data, err := repo.rowsToDomain(rows)
	if err != nil {
		return []JobStatus{}, common.WrapError(err)
	}
	return data, nil
}

// GetByJobIdBusinessDate retrieves JobStatus structs for a specific job id and business date.
//
// Mutates receiver: no
func (repo dbSqlPgRepo) GetByJobIdBusinessDate(jobId JobIdType, busDt common.Date) ([]JobStatus, error) {
	rows, err := repo.db.Query(repo.sqlSelect+repo.sqlWhereJobIdBusinessDate, jobId, time.Time(busDt))
	if err != nil {
		code := common.PgErrToCommon(err)
		return []JobStatus{}, common.NewCommonError(err, code, map[string]any{"jobId": jobId, "busDt": busDt})
	}
	defer rows.Close()

	data, err := repo.rowsToDomain(rows)
	if err != nil {
		return []JobStatus{}, common.WrapError(err)
	}
	return data, nil
}

// rowsToDomain converts a slice of database job status data to a slice of domain data by calling dbToDomain() for each item.
// If dbToDomain() fails to convert any row in the result set, it returns an empty slice and an error.
//
// Mutates receiver: no (doesn't use; receiver for namespace only)
func (repo dbSqlPgRepo) rowsToDomain(rows *sql.Rows) ([]JobStatus, error) {
	var result []JobStatus

	for rows.Next() {

		jobStatus, err := repo.dbToDomain(rows)
		if err != nil {
			return []JobStatus{}, common.WrapError(err)
		}

		result = append(result, jobStatus)
	}
	return result, nil
}

// dbToDomain converts database job status data to a JobStatus struct by scanning rows for values and building JobStatus.
//
// Mutates receiver: no (doesn't use; receiver for namespace only)
func (repo dbSqlPgRepo) dbToDomain(rows *sql.Rows) (JobStatus, error) {
	var (
		appId string
		jobId JobIdType
		jobSt JobStatusCodeType
		jobTs time.Time
		busDt time.Time // database/sql will Scan to time.Time, not common.Date
		runId string
		hstId string
	)

	err := rows.Scan(&appId, &jobId, &jobSt, &jobTs, &busDt, &runId, &hstId)
	if err != nil {
		return JobStatus{}, common.NewCommonError(err, common.ErrcdRepoScan, rows)
	}

	return newJobStatus(JobStatusDto{
		AppId: appId,
		JobId: string(jobId),
		JobSt: string(jobSt),
		JobTs: jobTs,
		BusDt: common.NewDateFromTime(busDt),
		RunId: runId,
		HstId: hstId,
	})
}

// domainToDb converts a JobStatus into an array of values to insert.
// SQL statements that specify values must use the expected order.
//
// Expected order: ApplicationId, JobId, JobStatusCode, BusinessDate, RunId, HostId
//
// Mutates receiver: no (doesn't use; receiver for namespace only)
func (db dbSqlPgRepo) domainToDb(jobStatus JobStatus) []any {
	return []any{
		jobStatus.ApplicationId,
		jobStatus.JobId,
		jobStatus.JobStatusCode,
		jobStatus.JobStatusTimestamp,
		time.Time(jobStatus.BusinessDate),
		jobStatus.RunId,
		jobStatus.HostId,
	}
}
