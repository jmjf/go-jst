package jobStatus

import (
	"time"
)

// I'm doing status this way now.
// If I want to have database table of job statuses, I could build a map[string]int and load it.
// That's overkill for now, so keeping it simple.
type JobStatusCodeType string
type JobIdType string

const (
	JobStatus_START   JobStatusCodeType = "Start"
	JobStatus_SUCCEED JobStatusCodeType = "Succeed"
	JobStatus_FAIL    JobStatusCodeType = "Fail"
)

type JobStatus struct {
	ApplicationId      string
	JobId              JobIdType
	JobStatusCode      JobStatusCodeType
	JobStatusTimestamp time.Time
	BusinessDate       time.Time
	RunId              string
	HostId             string
}

// Need to use uppercase field names so reflect can see them to use tags
type JobStatusDto struct {
	AppId string            `json:"applicationId"`
	JobId JobIdType         `json:"jobId"`
	JobSt JobStatusCodeType `json:"jobStatusCode"`
	JobTs time.Time         `json:"jobStatusTimestamp"`
	BusDt time.Time         `json:"businessDate"`
	RunId string            `json:"runId"`
	HstId string            `json:"hostId"`
}

type JobStatusRepo interface {
	add(jobStatus JobStatus) error
	GetByJobId(id JobIdType) ([]JobStatus, error)
	GetByJobIdBusinessDate(id JobIdType, businessDate time.Time) ([]JobStatus, error)
}

type JobStatusUC interface {
	Add(dto JobStatusDto) (JobStatus, error)
	GetByJobId(dto JobStatusDto) ([]JobStatus, error)
	GetByJobIdBusDate(dto JobStatusDto)
}
