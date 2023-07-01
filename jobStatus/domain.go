package jobStatus

import (
	"common"
	"time"
)

type JobStatus struct {
	ApplicationId      string            `json:"applicationId"`
	JobId              JobIdType         `json:"jobId"`
	JobStatusCode      JobStatusCodeType `json:"jobStatusCode"`
	JobStatusTimestamp time.Time         `json:"jobStatusTimestamp"`
	BusinessDate       common.Date       `json:"businessDate"`
	RunId              string            `json:"runId"`
	HostId             string            `json:"hostId"`
}

// newJobStatus validates the DTO and returns a new JobStatus using data from the DTO.
// This function should not be called outside the JobStatus package, so it is not exported.
//
// If the DTO contains invalid data, it returns an error with details.
func newJobStatus(dto JobStatusDto) (JobStatus, error) {
	errs := dto.isUsable()
	if len(errs) > 0 {
		return JobStatus{}, common.NewCommonError(common.ErrDomainProps, common.ErrcdDomainProps, errs)
	}

	// dto.isUsable() will return an error if the job status code isn't valid
	// so, here, dto.jobStatusCode() will return a valid value (not JobStatus_INVALID).
	return JobStatus{
		ApplicationId:      dto.AppId,
		JobId:              JobIdType(dto.JobId),
		JobStatusCode:      dto.jobStatusCode(),
		JobStatusTimestamp: common.TruncateTimeToMs(dto.JobTs),
		BusinessDate:       dto.BusDt,
		RunId:              dto.RunId,
		HostId:             dto.HstId,
	}, nil
}
