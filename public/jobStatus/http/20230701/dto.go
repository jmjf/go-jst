package dtoType

import (
	"go-slo/internal"
	"time"
)

type JobStatusDto struct {
	AppId string        `json:"applicationId"`
	JobId string        `json:"jobId"`
	JobSt string        `json:"jobStatusCode"`
	JobTs time.Time     `json:"jobStatusTimestamp"`
	BusDt internal.Date `json:"businessDate"`
	RunId string        `json:"runId"`
	HstId string        `json:"hostId"`
}
