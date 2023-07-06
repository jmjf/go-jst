package jobStatus

import (
	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

type jobStatusUC struct {
	jobStatusRepo JobStatusRepo
}

// NewJobStatusUC creates and returns a jobStatusUc.
func NewJobStatusUC(jsr JobStatusRepo) JobStatusUC {
	return &jobStatusUC{
		jobStatusRepo: jsr,
	}
}

// jobStatusUc.Add attempts to add a new job status to the data store.
// Returns a JobStatus and nil error on success.
// Returns an error (CommonError) and empty JobStatus on failure.
//
// Mutates receiver: no
func (uc jobStatusUC) Add(dto dtoType.JobStatusDto) (JobStatus, error) {
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

// TODO: jobStatusUc.GetByJobId
// Returns a slice of JobStatus and nil error on success.
// Returns an error (CommonError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc jobStatusUC) GetByJobId(dto dtoType.JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}

// TODO: jobStatusUc.GetByJobIdBusDt
// Returns a slice of JobStatus and nil error on success.
// Returns an error (CommonError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc jobStatusUC) GetByJobIdBusDt(dto dtoType.JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}
