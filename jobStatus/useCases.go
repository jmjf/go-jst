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

func NewJobStatusUC(jsr JobStatusRepo) JobStatusUC {
	return &jobStatusUC{
		jobStatusRepo: jsr,
	}
}

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

func (uc jobStatusUC) GetByJobId(dto JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}

func (uc jobStatusUC) GetByJobIdBusDt(dto JobStatusDto) ([]JobStatus, error) {
	return []JobStatus{}, nil
}
