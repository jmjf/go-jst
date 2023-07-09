package jobStatus

import (
	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

// AddJobStatusUC holds a Repo for it's Execute() method
type AddJobStatusUC struct {
	jobStatusRepo Repo
}

// NewJobStatusUC creates and returns an AddJobStatusUC.
func NewAddJobStatusUC(jsr Repo) *AddJobStatusUC {
	return &AddJobStatusUC{
		jobStatusRepo: jsr,
	}
}

// AddJobStatusUC.Execute attempts to add a new job status to the data store.
// Returns a JobStatus and nil error on success.
// Returns an error (CommonError) and empty JobStatus on failure.
//
// Mutates receiver: no
func (uc AddJobStatusUC) Execute(dto dtoType.JobStatusDto) (JobStatus, error) {
	jobStatus, err := NewJobStatus(dto)
	if err != nil {
		return JobStatus{}, internal.WrapError(err)
	}

	err = uc.jobStatusRepo.Add(jobStatus)
	if err != nil {
		return JobStatus{}, internal.WrapError(err)
	}

	return jobStatus, nil
}

type GetJobStatusByQueryUC struct {
	jobStatusRepo Repo
}

// NewGetJobStatusByQueryUC creates and returns a GetJobStatusByQueryUC
func NewGetJobStatusByQueryUC(jsr Repo) *GetJobStatusByQueryUC {
	return &GetJobStatusByQueryUC{
		jobStatusRepo: jsr,
	}
}

// TODO: GetJobStatusByQueryUC.Execute finds job status data for a query (string).
// Returns a slice of JobStatus and nil error on success.
// Returns an error (CommonError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc GetJobStatusByQueryUC) Execute(dto dtoType.JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}
