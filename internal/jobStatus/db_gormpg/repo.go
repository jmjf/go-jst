package repo

import (
	"time"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	dtoType "go-slo/public/jobStatus/http/20230701"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type repoDB struct {
	DSN string
	DB  *gorm.DB
}

type gormModel struct {
	ApplicationId      string    `gorm:"primaryKey;column:ApplicationId"`
	JobId              string    `gorm:"primaryKey;column:JobId"`
	JobStatusCode      string    `gorm:"column:JobStatusCode"`
	JobStatusTimestamp time.Time `gorm:"primaryKey;column:JobStatusTimestamp"`
	BusinessDate       time.Time `gorm:"primaryKey;column:BusinessDate"`
	RunId              string    `gorm:"column:RunId"`
	HostId             string    `gorm:"column:HostId"`
}

func (gormModel) TableName() string {
	return "JobStatus"
}

// NewRepoDb creates a new database/ORM specific object using the passed database handle.
// Passing the handle lets it be setup during application startup and shared with other repos.
func NewRepoDB(DSN string) *repoDB {
	return &repoDB{
		DSN: DSN,
	}
}

// Open connects to the database described by the dsn set on the repo.
//
// Mutates receiver: yes (sets repo.DB)
func (repo *repoDB) Open() error {
	if repo.DSN == "" {
		return internal.NewLoggableError(internal.ErrRepoNoDsn, internal.ErrcdRepoNoDsn, nil)
	}

	db, err := gorm.Open(postgres.Open(repo.DSN), &gorm.Config{
		TranslateError: false,
		Logger:         gormLogger.Default.LogMode(gormLogger.Silent),
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
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
	// gorm uses a connection pool, so doesn't have a direct Close()
	// get the sql.DB it's using and close it because I'm opening pools per repo
	// need to think about how much sense that makes
	if repo.DB != nil {
		sqlDB, err := repo.DB.DB()
		if err != nil {
			return internal.NewLoggableError(err, internal.PgErrToCommon((err)), nil)
		}
		return sqlDB.Close()
	}
	return nil
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: no
func (repo *repoDB) Add(jobStatus jobStatus.JobStatus) error {
	// we only care that it succeeds, not looking for a return, so use Exec()
	data := domainToDb(jobStatus)
	res := repo.DB.Create(&data)

	if res.Error != nil {
		code := internal.PgErrToCommon(res.Error)
		return internal.NewLoggableError(res.Error, code, jobStatus)
	}

	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: no
func (repo *repoDB) GetByJobId(jobId jobStatus.JobIdType) ([]jobStatus.JobStatus, error) {
	var dbStatuses []gormModel
	where := map[string]string{"jobId": string(jobId)}

	// Use named argument to avoid questions about tags
	result := repo.DB.Where("JobId = @jobId", where).Find(&dbStatuses)

	if result.Error != nil {
		code := internal.PgErrToCommon(result.Error)
		return nil, internal.NewLoggableError(result.Error, code, where)
	}

	data, err := rowsToDomain(dbStatuses)
	if err != nil {
		return nil, internal.WrapError(err)
	}
	return data, nil
}

// GetByJobIdBusinessDate retrieves JobStatus structs for a specific job id and business date.
//
// Mutates receiver: no
func (repo *repoDB) GetByJobIdBusinessDate(jobId jobStatus.JobIdType, busDt internal.Date) ([]jobStatus.JobStatus, error) {
	var dbStatuses []gormModel
	where := map[string]any{
		"jobId": jobId,
		"busDt": time.Time(busDt),
	}

	// Use named argument to avoid questions about tags
	result := repo.DB.Where("JobId = @jobId and BusinessDate = @busDt", where).Find(&dbStatuses)

	if result.Error != nil {
		code := internal.PgErrToCommon(result.Error)
		return nil, internal.NewLoggableError(result.Error, code, where)
	}

	data, err := rowsToDomain(dbStatuses)
	if err != nil {
		return nil, internal.WrapError(err)
	}
	return data, nil
}

// domainToGormPg converts a JobStatus to a GormJobStatusModel
func domainToDb(jobStatus jobStatus.JobStatus) gormModel {
	return gormModel{
		ApplicationId:      jobStatus.ApplicationId,
		JobId:              string(jobStatus.JobId),
		JobStatusCode:      string(jobStatus.JobStatusCode),
		JobStatusTimestamp: jobStatus.JobStatusTimestamp,
		BusinessDate:       jobStatus.BusinessDate.AsTime(),
		RunId:              jobStatus.RunId,
		HostId:             jobStatus.HostId,
	}
}

// rowsToDomain converts a slice of database job status data to a slice of domain data by calling dbToDomain() for each item.
// If dbToDomain() fails to convert any row in the result set, it returns an empty slice and an error.
func rowsToDomain(dbStatuses []gormModel) ([]jobStatus.JobStatus, error) {
	var result []jobStatus.JobStatus

	for _, dbStatus := range dbStatuses {

		jobStatus, err := dbToDomain(dbStatus)
		if err != nil {
			return nil, internal.WrapError(err)
		}

		result = append(result, jobStatus)
	}
	return result, nil
}

// dbToDomain converts one database job status to a JobStatus by calling newJobStatus().
func dbToDomain(dbStatus gormModel) (jobStatus.JobStatus, error) {
	return jobStatus.NewJobStatus(dtoType.JobStatusDto{
		AppId: dbStatus.ApplicationId,
		JobId: dbStatus.JobId,
		JobSt: dbStatus.JobStatusCode,
		JobTs: dbStatus.JobStatusTimestamp,
		BusDt: internal.NewDateFromTime(dbStatus.BusinessDate),
		RunId: dbStatus.RunId,
		HstId: dbStatus.HostId,
	})
}
