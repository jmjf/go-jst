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

type repoDb struct {
	DSN string
	DB  *gorm.DB
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

// NewRepoDb creates a new database/ORM specific object using the passed database handle.
// Passing the handle lets it be setup during application startup and shared with other repos.
func NewRepoDb(DSN string) *repoDb {
	return &repoDb{
		DSN: DSN,
	}
}

// Open connects to the database described by the dsn set on the repo.
//
// Mutates receiver: yes (sets repo.DB)
func (repo *repoDb) Open() error {
	if repo.DSN == "" {
		return internal.NewCommonError(internal.ErrRepoNoDsn, internal.ErrcdRepoNoDsn, nil)
	}

	db, err := gorm.Open(postgres.Open(repo.DSN), &gorm.Config{
		TranslateError: false, // get raw Postgres errors because they're more expressive
		Logger:         gormLogger.Default.LogMode(gormLogger.Silent),
		// Logger: logger, // doesn't work because gorm's logger interface is different; will need to translate
		NowFunc: func() time.Time { return time.Now().UTC() }, // ensure times are UTC
		// PrepareStmt: true // cache prepared statements for SQL; need to investigate how this works before turning on
	})
	if err != nil {
		return internal.NewCommonError(err, internal.ErrcdRepoConnException, nil)
	}
	repo.DB = db
	return nil
}

// Close closes the repo's database connection
//
// Mutates receiver: no
func (repo *repoDb) Close() error {
	// gorm uses a connection pool, so doesn't have a direct Close()
	// get the sql.DB it's using and close it because I'm opening pools per repo
	// need to think about how much sense that makes
	if repo.DB != nil {
		sqlDB, err := repo.DB.DB()
		if err != nil {
			return internal.NewCommonError(err, internal.PgErrToCommon((err)), nil)
		}
		return sqlDB.Close()
	}
	return nil
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: no
func (repo *repoDb) Add(jobStatus jobStatus.JobStatus) error {
	// we only care that it succeeds, not looking for a return, so use Exec()
	dbData := domainToDb(jobStatus)
	result := repo.DB.Create(&dbData)

	if result.Error != nil {
		code := internal.PgErrToCommon(result.Error)
		return internal.NewCommonError(result.Error, code, jobStatus)
	}

	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: no
func (repo *repoDb) GetByJobId(jobId jobStatus.JobIdType) ([]jobStatus.JobStatus, error) {
	var dbStatuses []GormPgJobStatusModel
	whereMap := map[string]string{"jobId": string(jobId)}

	// Use named argument to avoid questions about tags
	result := repo.DB.Where("JobId = @jobId", whereMap).Find(&dbStatuses)

	if result.Error != nil {
		code := internal.PgErrToCommon(result.Error)
		return nil, internal.NewCommonError(result.Error, code, map[string]any{"jobId": jobId})
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
func (repo *repoDb) GetByJobIdBusinessDate(jobId jobStatus.JobIdType, busDt internal.Date) ([]jobStatus.JobStatus, error) {
	var dbStatuses []GormPgJobStatusModel
	whereMap := map[string]any{
		"jobId": jobId,
		"busDt": time.Time(busDt),
	}

	// Use named argument to avoid questions about tags
	result := repo.DB.Where("JobId = @jobId and BusinessDate = @busDt", whereMap).Find(&dbStatuses)

	if result.Error != nil {
		code := internal.PgErrToCommon(result.Error)
		return nil, internal.NewCommonError(result.Error, code, map[string]any{"jobId": jobId, "busDt": busDt})
	}

	data, err := rowsToDomain(dbStatuses)
	if err != nil {
		return nil, internal.WrapError(err)
	}
	return data, nil
}

// domainToGormPg converts a JobStatus to a GormJobStatusModel
func domainToDb(jobStatus jobStatus.JobStatus) GormPgJobStatusModel {
	return GormPgJobStatusModel{
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
func rowsToDomain(dbStatuses []GormPgJobStatusModel) ([]jobStatus.JobStatus, error) {
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
func dbToDomain(dbStatus GormPgJobStatusModel) (jobStatus.JobStatus, error) {
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
