package jobStatus_test

import (
	"common"
	"database/sql"
	"errors"
	"fmt"
	"jobStatus"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func beforeEach(t *testing.T) (*sql.DB, sqlmock.Sqlmock, jobStatus.JobStatusUC, jobStatus.JobStatusDto, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	jsRepo := jobStatus.NewDbSqlPgRepo(db)
	uc := jobStatus.NewJobStatusUC(jsRepo)

	busDt, err := common.NewDate("2023-06-20")

	dto := jobStatus.JobStatusDto{
		AppId: "App1",
		JobId: "Job2",
		JobSt: string(jobStatus.JobStatus_START),
		JobTs: time.Now(),
		BusDt: busDt,
		RunId: "Run3",
		HstId: "Host4",
	}

	return db, mock, uc, dto, err
}

func Test_jobStatusUC_Add_InvalidDtoDataReturnsError(t *testing.T) {
	// value for BusDt test needs to be Date type
	futureDate, _ := common.NewDate(time.Now().Add(48 * time.Hour).Format(time.DateOnly))

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
			_, _, uc, dto, err := beforeEach(t) // don't need db or mock
			if err != nil {
				t.Fatal(err)
			}

			// set the value of the field to test
			tf := reflect.ValueOf(&dto).Elem().FieldByName(tt.testField)
			tf.Set(reflect.ValueOf(tt.testValue))

			// Act
			got, err := uc.Add(dto)

			if err == nil {
				t.Errorf("FAIL | Expected error %q, got: %+v", tt.wantErr, got)
				return
			}

			var de *common.DomainError
			if errors.As(err, &de) {
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

func Test_jobStatusUC_Add_DatabaseErrorReturnsError(t *testing.T) {
	// when it gets a database error, it should return an error

	// Arrange
	db, mock, uc, dto, err := beforeEach(t)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`INSERT INTO "JobStatus"`).
		WillReturnError(fmt.Errorf("fake database error"))

		// Act
	js, err := uc.Add(dto)

	// Assert
	if err == nil {
		t.Errorf("FAIL | Expected error, got err: %s  js: %+v", err, js)
		return
	}
	var re *common.RepoError
	if errors.As(err, &re) {
		if re.Code != "RepoOtherError" { // not exported, so using literal for now
			t.Errorf("FAIL | Expected RepoOtherError, got %+v", re)
		}
		// whether Code is wrong or not, we go the right type of error so we're done
		return
	}
	t.Errorf("FAIL | Expected RepoError, got err: %v", err)
}

func Test_jobStatusUC_Add_SuccessReturnsJobStatus(t *testing.T) {
	// when data is good it returns a JobStatus

	// Arrange
	db, mock, uc, dto, err := beforeEach(t)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`INSERT INTO "JobStatus"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	js, err := uc.Add(dto)

	// Assert
	if err != nil {
		t.Errorf("FAIL | Expected ok, got err: %+v", err)
		return
	}

	// extra safety checks for time data normalized; should never hit these

	// no longer needed with Date type
	// if tz, _ := js.BusinessDate.Zone(); tz != "UTC" || js.BusinessDate.Hour() != 0 ||
	// 	js.BusinessDate.Minute() != 0 || js.BusinessDate.Second() != 0 || js.BusinessDate.Nanosecond() != 0 {
	// 	t.Errorf("FAIL | BusinessDate not normalized %s", js.BusinessDate)
	// 	return
	// }
	if tz, _ := js.JobStatusTimestamp.Zone(); tz != "UTC" || js.JobStatusTimestamp.Nanosecond() != 0 {
		t.Errorf("FAIL | JobStatusTimestamp not normalized %s", js.JobStatusTimestamp)
	}
}

/***
func Test_jobStatusUC_Add(t *testing.T) {
	type fields struct {
		jobStatusRepo JobStatusRepo
	}
	type args struct {
		dto JobStatusDto
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    JobStatus
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := jobStatusUC{
				jobStatusRepo: tt.fields.jobStatusRepo,
			}
			got, err := uc.Add(tt.args.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("jobStatusUC.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jobStatusUC.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
***/
