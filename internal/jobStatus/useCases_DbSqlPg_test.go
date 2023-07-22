package jobStatus_test

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"testing"
	"time"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	repo "go-slo/internal/jobStatus/db_sqlpgx"
	dtoType "go-slo/public/jobStatus/http/20230701"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func dbSqlPgBeforeEach(t *testing.T) (*sql.DB, sqlmock.Sqlmock, jobStatus.Repo, dtoType.JobStatusDto, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	jsRepo := repo.NewRepoDB("")
	jsRepo.DB = db

	busDt, err := internal.NewDate("2023-06-20")

	dto := dtoType.JobStatusDto{
		AppId: "App1",
		JobId: "Job2",
		JobSt: string(jobStatus.JobStart),
		JobTs: internal.TruncateTimeToMs(time.Now()),
		BusDt: busDt,
		RunId: "Run3",
		HstId: "Host4",
	}

	return db, mock, jsRepo, dto, err
}

func Test_jobStatusUC_dbsqlpg_Add_InvalidDtoDataReturnsError(t *testing.T) {
	// value for BusDt test needs to be Date type
	futureDate, _ := internal.NewDate(time.Now().Add(48 * time.Hour).Format(time.DateOnly))

	tests := []struct {
		name      string
		testField string
		testValue any
		wantErr   string
	}{
		{
			name:      "when AppId is too short it returns invalid ApplicationId",
			testField: "AppId",
			testValue: "",
			wantErr:   "invalid ApplicationId",
		},
		{
			name:      "when AppId is too long it returns invalid ApplicationId",
			testField: "AppId",
			testValue: fmt.Sprintf("%201s", "a"),
			wantErr:   "invalid ApplicationId",
		},
		{
			name:      "when JobId is too short it returns invalid JobId",
			testField: "JobId",
			testValue: "",
			wantErr:   "invalid JobId",
		},
		{
			name:      "when JobId is too long it returns invalid JobId",
			testField: "JobId",
			testValue: fmt.Sprintf("%201s", "a"),
			wantErr:   "invalid JobId",
		},
		{
			name:      "when JobSt is too short it returns invalid JobStatusCode",
			testField: "JobSt",
			testValue: "",
			wantErr:   "invalid JobStatusCode",
		},
		{
			name:      "when JobSt is not a valid value it returns invalid JobStatusCode",
			testField: "JobSt",
			testValue: "random garbage",
			wantErr:   "invalid JobStatusCode",
		},
		{
			name:      "when RunId is to long it returns RunId over 50",
			testField: "RunId",
			testValue: fmt.Sprintf("%51s", "a"),
			wantErr:   "RunId is over 50 characters",
		},
		{
			name:      "when HstId is to long it returns HostId over 150",
			testField: "HstId",
			testValue: fmt.Sprintf("%151s", "a"),
			wantErr:   "HostId is over 150 characters",
		},
		{
			name:      "when JobTs is future it returns invalid JobTimestamp",
			testField: "JobTs",
			testValue: time.Now().Add(1 * time.Minute),
			wantErr:   "invalid JobTimestamp",
		},
		{
			name:      "when BusDt is future it returns invalid BusinessDate",
			testField: "BusDt",
			testValue: futureDate,
			wantErr:   "invalid BusinessDate",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			_, _, jsRepo, dto, err := dbSqlPgBeforeEach(t) // don't need db or mock
			if err != nil {
				t.Fatal(err)
			}

			uc := jobStatus.NewAddJobStatusUC(jsRepo)

			// set the value of the field to test
			tf := reflect.ValueOf(&dto).Elem().FieldByName(tt.testField)
			tf.Set(reflect.ValueOf(tt.testValue))

			// Act
			got, err := uc.Execute(dto)

			if err == nil {
				t.Errorf("FAIL | Expected error %q, got: %+v", tt.wantErr, got)
				return
			}

			var le *internal.LoggableError
			if errors.As(err, &le) {
				// get the first error from Data and call Error() on it to get a string
				msg := le.Data.([]error)[0].Error()
				match, _ := regexp.MatchString(tt.wantErr, msg)
				if !match {
					t.Errorf("FAIL | Expected error %q, got: %s", tt.wantErr, err)
				}
				// err is a LoggableError so, we're good
				return
			}
			t.Errorf("FAIL | Expected LoggableError, got: %v", err)
		})
	}
}

func Test_jobStatusUC_dbsqlpg_Add_RepoErrors(t *testing.T) {
	// when repo returns <error> it recognizes the error

	// Arrange
	tests := []struct {
		name          string
		testErr       error
		expectErrCode string
	}{
		{
			name:          "when repo returns RepoDupeRowError it recognizes the error",
			testErr:       &pgconn.PgError{Code: "23505"},
			expectErrCode: internal.ErrcdRepoDupeRow,
		},
		{
			name:          "when repo returns RepoConnExceptionError it recognizes the error",
			testErr:       &pgconn.PgError{Code: "08xxx"}, // a family of errors that begin with "08"
			expectErrCode: internal.ErrcdRepoConnException,
		},
		{
			name:          "when repo returns RepoOtherError it recognizes the error",
			testErr:       &pgconn.PgError{Code: "unknown"},
			expectErrCode: internal.ErrcdRepoOther,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			db, mock, jsRepo, dto, err := dbSqlPgBeforeEach(t)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			uc := jobStatus.NewAddJobStatusUC(jsRepo)

			mock.ExpectExec(`INSERT INTO "JobStatus"`).
				WithArgs(dto.AppId, dto.JobId, dto.JobSt, internal.MatchTime{Value: dto.JobTs}, internal.MatchTime{Value: time.Time(dto.BusDt)}, dto.RunId, dto.HstId).
				WillReturnError(tt.testErr)

				// Act
			js, err := uc.Execute(dto)

			// Assert
			if err == nil {
				t.Errorf("FAIL | Expected error, got err: %s  js: %+v", err, js)
				return
			}
			var le *internal.LoggableError
			if errors.As(err, &le) {
				// fmt.Printf("le %+v", *le)
				if le.Code != tt.expectErrCode {
					t.Errorf("FAIL | Expected %s, got %+v", tt.expectErrCode, le)
				}
				// whether Code is wrong or not, we go the right type of error so we're done
				return
			}
			t.Errorf("FAIL | Expected LoggableError, got err: %v", err)
		})
	}
}

func Test_jobStatusUC_dbsqlpg_Add_SuccessReturnsJobStatus(t *testing.T) {
	// when data is good it returns a JobStatus

	// Arrange
	db, mock, jsRepo, dto, err := dbSqlPgBeforeEach(t)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	uc := jobStatus.NewAddJobStatusUC(jsRepo)

	mock.ExpectExec(`INSERT INTO "JobStatus"`).
		WithArgs(dto.AppId, dto.JobId, dto.JobSt, internal.MatchTime{Value: dto.JobTs}, internal.MatchTime{Value: time.Time(dto.BusDt)}, dto.RunId, dto.HstId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	_, err = uc.Execute(dto)

	// Assert
	if err != nil {
		t.Errorf("FAIL | Expected ok, got err: %+v", err)
		return
	}
}

func Test_jobStatusUC_dbsqlpg_GetQuery_InvalidQueryTermReturnsError(t *testing.T) {

	tests := []struct {
		name    string
		testURL string
		wantErr string
	}{
		{
			name:    "when query is empty it returns missing query term",
			testURL: "/test",
			wantErr: internal.ErrAppTermMissing.Error(),
		},
		{
			name:    "when a time present but applicationId and jobId are empty it returns missing query term",
			testURL: "/test?businessDate=2023-01-01",
			wantErr: internal.ErrAppTermMissing.Error(),
		},
		{
			name:    "when an id present, but jobStatusTimestamp and businessDate are empty it returns missing query term",
			testURL: "/test?jobId=abc123",
			wantErr: internal.ErrAppTermMissing.Error(),
		},
		// we don't test success here because later tests will cover that by working
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			_, _, jsRepo, _, err := dbSqlPgBeforeEach(t) // don't need db or mock
			if err != nil {
				t.Fatal(err)
			}

			uc := jobStatus.NewGetByQueryUC(jsRepo)

			// set the value of the field to test
			u, _ := url.Parse(tt.testURL)

			// Act
			got, err := uc.Execute(u.Query())
			if err == nil {
				t.Errorf("FAIL | Expected error %q, got: %+v", tt.wantErr, got)
				return
			}

			// Assert
			var le *internal.LoggableError
			if errors.As(err, &le) {
				// is the expected error string present
				match, _ := regexp.MatchString(tt.wantErr, le.Error())
				if !match {
					t.Errorf("FAIL | Expected error %q, got: %s", tt.wantErr, err)
				}
				return
			}
			t.Errorf("FAIL | Expected LoggableError, got: %v", err)
		})
	}
}

func Test_jobStatusUC_dbsqlpg_GetQuery_RepoErrorReturnsError(t *testing.T) {
	// Arrange
	wantErr := internal.ErrRepoOther.Error()
	db, mock, jsRepo, _, err := dbSqlPgBeforeEach(t)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	uc := jobStatus.NewGetByQueryUC(jsRepo)
	u, _ := url.Parse("/test?jobId=abc123&businessDate=2023-01-01&applicationId=")

	mock.ExpectQuery(`SELECT "ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId" FROM "JobStatus"`).
		WillReturnError(internal.ErrRepoOther)

	// Act
	got, err := uc.Execute(u.Query())
	if err == nil {
		t.Errorf("FAIL | Expected error %s, got: %+v", wantErr, got)
		return
	}

	// Assert
	var le *internal.LoggableError
	if errors.As(err, &le) {
		// is the expected error string present
		match, _ := regexp.MatchString(wantErr, le.Error())
		if !match {
			t.Errorf("FAIL | Expected error %q, got: %s", wantErr, err)
		}
		return
	}
	t.Errorf("FAIL | Expected LoggableError, got: %v", err)
}

func Test_jobStatusUC_dbsqlpg_GetQuery_NoDataReturnsEmptyResult(t *testing.T) {
	// Arrange
	db, mock, jsRepo, _, err := dbSqlPgBeforeEach(t)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	uc := jobStatus.NewGetByQueryUC(jsRepo)
	// test repo handling of applicationId , jobStatusCode, and hostId
	u, _ := url.Parse("/test?jobId=abc123&businessDate=2023-01-01&applicationId=987xyz&jobStatusCode=START&hostId=93")

	mock.ExpectQuery(`SELECT "ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId" FROM "JobStatus"`).
		WillReturnRows(sqlmock.NewRows([]string{"ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId"}))
		// empty result

	// Act
	got, err := uc.Execute(u.Query())
	if err != nil {
		t.Errorf("FAIL | Expected empty result, got: %+v", err)
		return
	}

	if len(got) != 0 {
		t.Errorf("FAIL | Expected empty result, got: %+v", got)
	}
}

func Test_jobStatusUC_dbsqlpg_GetQuery_DataFoundReturnsResult(t *testing.T) {
	// Arrange
	db, mock, jsRepo, _, err := dbSqlPgBeforeEach(t)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	uc := jobStatus.NewGetByQueryUC(jsRepo)
	// test repo handling of jobStatusTimestamp and runId
	u, _ := url.Parse("/test?jobId=abc123&businessDate=2023-01-01&jobStatusTimestamp=2023-01-02T05:06:07Z&runId=43")

	dt, _ := internal.NewDate("2023-04-08")
	wantData := []jobStatus.JobStatus{
		{
			ApplicationId:      "abc",
			JobId:              "123",
			JobStatusCode:      jobStatus.JobStart,
			JobStatusTimestamp: time.Now().UTC(),
			BusinessDate:       dt,
			RunId:              "run1",
			HostId:             "",
		},
		{
			ApplicationId:      "def",
			JobId:              "987",
			JobStatusCode:      jobStatus.JobFail,
			JobStatusTimestamp: time.Now().UTC(),
			BusinessDate:       dt,
			RunId:              "run2",
			HostId:             "42",
		},
	}
	rows := sqlmock.NewRows([]string{"ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId"}).
		AddRow(wantData[0].ApplicationId, wantData[0].JobId, wantData[0].JobStatusCode, wantData[0].JobStatusTimestamp,
			wantData[0].BusinessDate.AsTime(), wantData[0].RunId, wantData[0].HostId).
		AddRow(wantData[1].ApplicationId, wantData[1].JobId, wantData[1].JobStatusCode, wantData[1].JobStatusTimestamp,
			wantData[1].BusinessDate.AsTime(), wantData[1].RunId, wantData[1].HostId)

	mock.ExpectQuery(`SELECT "ApplicationId", "JobId", "JobStatusCode", "JobStatusTimestamp", "BusinessDate", "RunId", "HostId" FROM "JobStatus"`).
		WillReturnRows(rows)

	// Act
	got, err := uc.Execute(u.Query())
	if err != nil {
		t.Errorf("FAIL | Expected empty result, got: %+v", err)
		return
	}

	if len(got) != 2 {
		t.Errorf("FAIL | Expected 2 results, got: %d - %+v", len(got), got)
	}
	if got[0].ApplicationId != wantData[0].ApplicationId ||
		got[0].JobId != wantData[0].JobId ||
		got[1].ApplicationId != wantData[1].ApplicationId ||
		got[1].JobId != wantData[1].JobId {
		t.Errorf("FAIL | got doesn't match wantData\ngot: %+v\nwantData: %+v", got, wantData)
	}
}
