package jobStatus

import (
	"common"
	"time"

	"gorm.io/gorm"
)

type gormPgRepo struct {
	db *gorm.DB
}

type GormPgJobStatusModel struct {
	ApplicationId      string    `gorm:"primaryKey;column:ApplicationId"`
	JobId              string    `gorm:"primaryKey;column:JobId"`
	JobStatusCode      string    `gorm:"column:JobStatusCode"`
	JobStatusTimestamp time.Time `gorm:"primaryKey;column:JobStatusTimestamp"`
	BusinessDate       time.Time `gorm:"primaryKey;column:BusinessDate"`
	RunId              string    `gorm:"column:RunId"`
	HostId             string    `gorm:"column:HostId"`
}

func (GormPgJobStatusModel) TableName() string {
	return "JobStatus"
}

// NewGormPgRepo creates a new dbSqlRepo object using the passed database handle.
// Passing the handle lets it be setup during application startup and shared with other repos.
func NewGormPgRepo(db *gorm.DB) JobStatusRepo {
	return &gormPgRepo{
		db: db,
	}
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: no
func (repo gormPgRepo) add(jobStatus JobStatus) error {
	// we only care that it succeeds, not looking for a return, so use Exec()
	dbData := repo.domainToDb(jobStatus)
	result := repo.db.Create(&dbData)

	if result.Error != nil {
		code := common.PgErrToCommon(result.Error)
		return common.NewCommonError(result.Error, code, jobStatus)
	}

	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: no
func (repo gormPgRepo) GetByJobId(jobId JobIdType) ([]JobStatus, error) {
	var dbStatuses []GormPgJobStatusModel
	whereMap := map[string]string{"jobId": string(jobId)}

	// Use named argument to avoid questions about tags
	result := repo.db.Where("JobId = @jobId", whereMap).Find(&dbStatuses)

	if result.Error != nil {
		code := common.PgErrToCommon(result.Error)
		return []JobStatus{}, common.NewCommonError(result.Error, code, map[string]any{"jobId": jobId})
	}

	data, err := repo.rowsToDomain(dbStatuses)
	if err != nil {
		return []JobStatus{}, common.WrapError(err)
	}
	return data, nil
}

// GetByJobIdBusinessDate retrieves JobStatus structs for a specific job id and business date.
//
// Mutates receiver: no
func (repo gormPgRepo) GetByJobIdBusinessDate(jobId JobIdType, busDt common.Date) ([]JobStatus, error) {
	var dbStatuses []GormPgJobStatusModel
	whereMap := map[string]any{
		"jobId": jobId,
		"busDt": time.Time(busDt),
	}

	// Use named argument to avoid questions about tags
	result := repo.db.Where("JobId = @jobId and BusinessDate = @busDt", whereMap).Find(&dbStatuses)

	if result.Error != nil {
		code := common.PgErrToCommon(result.Error)
		return []JobStatus{}, common.NewCommonError(result.Error, code, map[string]any{"jobId": jobId, "busDt": busDt})
	}

	data, err := repo.rowsToDomain(dbStatuses)
	if err != nil {
		return []JobStatus{}, common.WrapError(err)
	}
	return data, nil
}

// domainToGormPg converts a JobStatus to a GormJobStatusModel
//
// Mutates receiver: no (doesn't use; receiver for namespace only)
func (db gormPgRepo) domainToDb(jobStatus JobStatus) GormPgJobStatusModel {
	return GormPgJobStatusModel{
		ApplicationId:      jobStatus.ApplicationId,
		JobId:              string(jobStatus.JobId),
		JobStatusCode:      string(jobStatus.JobStatusCode),
		JobStatusTimestamp: jobStatus.JobStatusTimestamp,
		BusinessDate:       time.Time(jobStatus.BusinessDate),
		RunId:              jobStatus.RunId,
		HostId:             jobStatus.HostId,
	}
}

// rowsToDomain converts a slice of database job status data to a slice of domain data by calling dbToDomain() for each item.
// If dbToDomain() fails to convert any row in the result set, it returns an empty slice and an error.
//
// Mutates receiver: no (doesn't use; receiver for namespace only)
func (repo gormPgRepo) rowsToDomain(dbStatuses []GormPgJobStatusModel) ([]JobStatus, error) {
	var result []JobStatus

	for _, dbStatus := range dbStatuses {

		jobStatus, err := repo.dbToDomain(dbStatus)
		if err != nil {
			return []JobStatus{}, common.WrapError(err)
		}

		result = append(result, jobStatus)
	}
	return result, nil
}

// dbToDomain converts one database job status to a JobStatus by calling newJobStatus().
//
// Mutates receiver: no (doesn't use; receiver for namespace only)
func (repo gormPgRepo) dbToDomain(dbStatus GormPgJobStatusModel) (JobStatus, error) {
	return newJobStatus(JobStatusDto{
		AppId: dbStatus.ApplicationId,
		JobId: dbStatus.JobId,
		JobSt: dbStatus.JobStatusCode,
		JobTs: dbStatus.JobStatusTimestamp,
		BusDt: common.NewDateFromTime(dbStatus.BusinessDate),
		RunId: dbStatus.RunId,
		HstId: dbStatus.HostId,
	})
}
