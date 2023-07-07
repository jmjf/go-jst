package jobStatus_test

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	repo "go-slo/internal/jobStatus/db/dbSqlPgx"
	dtoType "go-slo/public/jobStatus/http/20230701"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func dbSqlPgBeforeEach(t *testing.T) (*sql.DB, sqlmock.Sqlmock, jobStatus.JobStatusRepo, dtoType.JobStatusDto, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	jsRepo := repo.NewDbSqlPgRepo(db)

	busDt, err := internal.NewDate("2023-06-20")

	dto := dtoType.JobStatusDto{
		AppId: "App1",
		JobId: "Job2",
		JobSt: string(jobStatus.JobStatus_START),
		JobTs: internal.TruncateTimeToMs(time.Now()),
		BusDt: busDt,
		RunId: "Run3",
		HstId: "Host4",
	}

	return db, mock, jsRepo, dto, err
}

func Test_jobStatusUC_Add_InvalidDtoDataReturnsError(t *testing.T) {
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

			var de *internal.CommonError
			if errors.As(err, &de) {
				// get the first error from Data and call Error() on it to get a string
				msg := de.Data.([]error)[0].Error()
				match, _ := regexp.MatchString(tt.wantErr, msg)
				if !match {
					t.Errorf("FAIL | Expected error %q, got: %s", tt.wantErr, err)
				}
				// err is a DomainError so, we're good
				return
			}
			t.Errorf("FAIL | Expected DomainError, got: %v", err)

		})
	}
}

func Test_jobStatusUC_Add_RepoErrors(t *testing.T) {
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
			var re *internal.CommonError
			if errors.As(err, &re) {
				// fmt.Printf("re %+v", *re)
				if re.Code != tt.expectErrCode {
					t.Errorf("FAIL | Expected %s, got %+v", tt.expectErrCode, re)
				}
				// whether Code is wrong or not, we go the right type of error so we're done
				return
			}
			t.Errorf("FAIL | Expected CommonError, got err: %v", err)
		})
	}
}

func Test_jobStatusUC_Add_SuccessReturnsJobStatus(t *testing.T) {
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
