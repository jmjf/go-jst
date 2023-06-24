package jobStatus

import (
	"fmt"
	"strings"
	"time"
)

// Need to use uppercase field names so reflect can see them to use tags
type JobStatusDto struct {
	AppId string    `json:"applicationId"`
	JobId string    `json:"jobId"`
	JobSt string    `json:"jobStatusCode"`
	JobTs time.Time `json:"jobStatusTimestamp"`
	BusDt time.Time `json:"businessDate"`
	RunId string    `json:"runId"`
	HstId string    `json:"hostId"`
}

// isUsable checks data on the DTO to ensure it can be used to create a JobStatus.
//
// It returns an array of errors describing all data problems.
// If len(errs) == 0, the DTO is good.
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
	if now.Compare(dto.BusDt) == -1 {
		errs = append(errs, fmt.Errorf("invalid BusinessDate |%s|", dto.BusDt.Format(time.RFC3339)))
	}

	// if JobTs < BusDt -> error
	// need to think about this for TZ near international date line
	if dto.JobTs.Compare(dto.BusDt) == -1 {
		errs = append(errs, fmt.Errorf("JobTimestamp is less than BusinessDate |%s| |%s|", dto.JobTs.Format(time.RFC3339), dto.BusDt.Format(time.RFC3339)))
	}

	if len(dto.RunId) > 50 {
		errs = append(errs, fmt.Errorf("RunId is over 50 characters |%s|", dto.RunId))
	}

	if len(dto.HstId) > 150 {
		errs = append(errs, fmt.Errorf("HostId is over 150 characters |%s|", dto.HstId))
	}

	return errs
}

// normalizeTimes normalizes JobTs and BusDt to expected precision.
//
// Returns a JobStatusDto with JobTs truncated to seconds and converted to UTC
// and BusDt truncated to day as UTC (not converted to UTC).
func (dto *JobStatusDto) normalizeTimes() {
	dto.JobTs = dto.JobTs.Truncate(time.Second).UTC()

	yr, mo, dy := dto.BusDt.Date()
	dto.BusDt = time.Date(yr, mo, dy, 0, 0, 0, 0, time.UTC)
}

// jobStatusCode returns the job status code if dto.JobSt in the list of valid job status codes.
// If code is not a valid job status code, it returns JobStatus_INVALID.
func (dto *JobStatusDto) jobStatusCode() JobStatusCodeType {
	for _, jsc := range validJobStatusCodes {
		if string(jsc) == strings.ToUpper(dto.JobSt) {
			return jsc
		}
	}
	return JobStatus_INVALID
}

type JobStatus struct {
	ApplicationId      string
	JobId              JobIdType
	JobStatusCode      JobStatusCodeType
	JobStatusTimestamp time.Time
	BusinessDate       time.Time
	RunId              string
	HostId             string
}

// newJobStatus validates the DTO and returns a new JobStatus using data from the DTO.
// This function should not be called outside the JobStatus package, so it is not exported.
//
// If the DTO contains invalid data, it returns an error with details.
func newJobStatus(dto JobStatusDto) (JobStatus, error) {
	dto.normalizeTimes()

	errs := dto.isUsable()
	if len(errs) > 0 {
		// really want a custom error type and bundle all the errors in it
		return JobStatus{}, fmt.Errorf("DTO is not usable | %v", errs)
	}

	// dto.isUsable() will return an error if the job status code isn't valid
	// so, here, dto.jobStatusCode() will return a valid value (not JobStatus_INVALID).
	return JobStatus{
		ApplicationId:      dto.AppId,
		JobId:              JobIdType(dto.JobId),
		JobStatusCode:      dto.jobStatusCode(),
		JobStatusTimestamp: dto.JobTs,
		BusinessDate:       dto.BusDt,
		RunId:              dto.RunId,
		HostId:             dto.HstId,
	}, nil
}
