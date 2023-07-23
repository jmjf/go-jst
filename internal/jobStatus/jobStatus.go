package jobStatus

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

// interfaces used in jobStatus and subpackages
type Repo interface {
	// if running testRepo, change add() to Add() here and in the repos.
	Add(jobStatus JobStatus) error
	GetByQuery(map[string]string) ([]JobStatus, error)
	Close() error
}

// I'm doing status these types the easy way for now.
// If I want to have database table of job statuses, I could build a map[string]int and load it.
// That's overkill for now, so keeping it simple.
type JobStatusCodeType string
type JobIdType string

const (
	InvalidStatus JobStatusCodeType = "INVALID"
	JobStart      JobStatusCodeType = "START"
	JobSucceed    JobStatusCodeType = "SUCCEED"
	JobFail       JobStatusCodeType = "FAIL"
)

// Use validJobStatusCodes to ensure a value is a valid job status; update if job status consts change.
// INVALID is not a valid job status code, but it's defined for easy, safe checks for invalid.
var validJobStatusCodes = []JobStatusCodeType{JobStart, JobSucceed, JobFail}

type JobStatus struct {
	ApplicationId      string            `json:"applicationId"`
	JobId              JobIdType         `json:"jobId"`
	JobStatusCode      JobStatusCodeType `json:"jobStatusCode"`
	JobStatusTimestamp time.Time         `json:"jobStatusTimestamp"`
	BusinessDate       internal.Date     `json:"businessDate"`
	RunId              string            `json:"runId"`
	HostId             string            `json:"hostId"`
}

// structJSONMap takes a structure with JSON tags and builds a map of JSON tag name to field name.
// If s is a not a struct, it returns an empty map. If a field does not have a JSON tag, it skips the field.
// TODO: Find a better place for this function because it's probably useful elsewhere.
func structJSONMap(s any) map[string]string {
	var jsonMap = make(map[string]string)

	st := reflect.TypeOf(s)
	if st.Kind() != reflect.Struct {
		return jsonMap
	}

	for _, f := range reflect.VisibleFields(st) {
		jTag := f.Tag.Get("json")
		jName := strings.Split(jTag, ",")[0]
		if len(f.Name) == 0 || len(jName) == 0 {
			continue
		}
		jsonMap[jName] = f.Name
	}
	return jsonMap
}

// validFields supports selecting valid fields and only valid fields from a query string.
// Uses a function because requiring a manual sync with JobStatus leaves room for forgetting.
var validFields = structJSONMap(JobStatus{})

// newJobStatus validates the DTO and returns a new JobStatus using data from the DTO.
// This function should not be called outside the JobStatus package, so it is not exported.
//
// If the DTO contains invalid data, it returns an error with details.
func NewJobStatus(dto dtoType.JobStatusDto) (JobStatus, error) {
	errs := isDtoUsable(dto)
	if len(errs) > 0 {
		return JobStatus{}, internal.NewLoggableError(internal.ErrDomainProps, internal.ErrcdDomainProps, errs)
	}

	// dto.isUsable() will return an error if the job status code isn't valid
	// so, here, dto.jobStatusCode() will return a valid value (not InvalidStatus).
	return JobStatus{
		ApplicationId:      dto.AppId,
		JobId:              JobIdType(dto.JobId),
		JobStatusCode:      jobStatusCode(dto.JobSt),
		JobStatusTimestamp: internal.TruncateTimeToMs(dto.JobTs),
		BusinessDate:       dto.BusDt,
		RunId:              dto.RunId,
		HostId:             dto.HstId,
	}, nil
}

// isUsable checks data on the DTO to ensure it can be used to create a JobStatus.
// It returns an array of errors describing all data problems.
// If len(errs) == 0, the DTO is good.
func isDtoUsable(dto dtoType.JobStatusDto) []error {
	errs := []error{}
	now := time.Now()

	if len(dto.AppId) == 0 || len(dto.AppId) > 200 {
		errs = append(errs, fmt.Errorf("invalid ApplicationId |%s|", dto.AppId))
	}

	if len(dto.JobId) == 0 || len(dto.JobId) > 200 {
		errs = append(errs, fmt.Errorf("invalid JobId |%s|", dto.JobId))
	}

	if len(dto.JobSt) == 0 || jobStatusCode(dto.JobSt) == InvalidStatus {
		errs = append(errs, fmt.Errorf("invalid JobStatusCode |%s|", dto.JobSt))
	}

	// now < JobTs -> JobTs is in the future
	if now.Compare(dto.JobTs) == -1 {
		errs = append(errs, fmt.Errorf("invalid JobTimestamp |%s|", dto.JobTs.Format(time.RFC3339)))
	}

	// now < BusDt -> BusDt is in the future
	if now.Compare(time.Time(dto.BusDt)) == -1 {
		errs = append(errs, fmt.Errorf("invalid BusinessDate |%s|", dto.BusDt))
	}

	// if JobTs < BusDt -> error
	// need to think about this for TZ near international date line
	if dto.JobTs.Compare(time.Time(dto.BusDt)) == -1 {
		errs = append(errs, fmt.Errorf("JobTimestamp is less than BusinessDate |%s| |%s|", dto.JobTs.Format(time.RFC3339), dto.BusDt))
	}

	if len(dto.RunId) > 50 {
		errs = append(errs, fmt.Errorf("RunId is over 50 characters |%s|", dto.RunId))
	}

	if len(dto.HstId) > 150 {
		errs = append(errs, fmt.Errorf("HostId is over 150 characters |%s|", dto.HstId))
	}

	return errs
}

// jobStatusCode returns the job status code if code is in the list of valid job status codes.
// If code is not a valid job status code, it returns InvalidStatus.
func jobStatusCode(code string) JobStatusCodeType {
	for _, jsc := range validJobStatusCodes {
		if string(jsc) == strings.ToUpper(code) {
			return jsc
		}
	}
	return InvalidStatus
}
