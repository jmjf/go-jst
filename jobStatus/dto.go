package jobStatus

import (
	"common"
	"fmt"
	"strings"
	"time"
)

// Need to use uppercase field names so reflect can see them to use tags
type JobStatusDto struct {
	AppId string      `json:"applicationId"`
	JobId string      `json:"jobId"`
	JobSt string      `json:"jobStatusCode"`
	JobTs time.Time   `json:"jobStatusTimestamp"`
	BusDt common.Date `json:"businessDate"`
	RunId string      `json:"runId"`
	HstId string      `json:"hostId"`
}

// isUsable checks data on the DTO to ensure it can be used to create a JobStatus.
// It returns an array of errors describing all data problems.
// If len(errs) == 0, the DTO is good.
//
// Mutates receiver: no
func (dto JobStatusDto) isUsable() []error {
	errs := []error{}
	now := time.Now()

	if len(dto.AppId) == 0 || len(dto.AppId) > 200 {
		errs = append(errs, fmt.Errorf("invalid ApplicationId |%s|", dto.AppId))
	}

	if len(dto.JobId) == 0 || len(dto.JobId) > 200 {
		errs = append(errs, fmt.Errorf("invalid JobId |%s|", dto.JobId))
	}

	if len(dto.JobSt) == 0 || dto.jobStatusCode() == JobStatus_INVALID {
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

// jobStatusCode returns the job status code if dto.JobSt in the list of valid job status codes.
// If code is not a valid job status code, it returns JobStatus_INVALID.
//
// Mutates receiver: no
func (dto JobStatusDto) jobStatusCode() JobStatusCodeType {
	for _, jsc := range validJobStatusCodes {
		if string(jsc) == strings.ToUpper(dto.JobSt) {
			return jsc
		}
	}
	return JobStatus_INVALID
}
