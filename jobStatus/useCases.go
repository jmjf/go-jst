package jobStatus

import "common"

type JobStatusUC interface {
	Add(dto JobStatusDto) (JobStatus, error)
	GetByJobId(dto JobStatusDto) ([]JobStatus, error)
	GetByJobIdBusDt(dto JobStatusDto) ([]JobStatus, error)
}

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
func (uc jobStatusUC) Add(dto JobStatusDto) (JobStatus, error) {
	jobStatus, err := newJobStatus(dto)
	if err != nil {
		return JobStatus{}, common.WrapError(err)
	}

	err = uc.jobStatusRepo.add(jobStatus)
	if err != nil {
		return JobStatus{}, common.WrapError(err)
	}

	return jobStatus, nil
}

// TODO: jobStatusUc.GetByJobId
// Returns a slice of JobStatus and nil error on success.
// Returns an error (CommonError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc jobStatusUC) GetByJobId(dto JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}

// TODO: jobStatusUc.GetByJobIdBusDt
// Returns a slice of JobStatus and nil error on success.
// Returns an error (CommonError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc jobStatusUC) GetByJobIdBusDt(dto JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}
