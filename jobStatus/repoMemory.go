package jobStatus

import (
	"common"
	"sync"
)

type memoryRepo struct {
	jobStatuses []JobStatus
	mut         sync.Mutex
}

func NewMemoryRepo(jobStatuses []JobStatus) JobStatusRepo {
	return &memoryRepo{jobStatuses: jobStatuses}
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: yes (mutex, data)
func (repo *memoryRepo) add(jobStatus JobStatus) error {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	repo.jobStatuses = append(repo.jobStatuses, jobStatus)
	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: yes (mutex)
func (repo *memoryRepo) GetByJobId(jobId JobIdType) ([]JobStatus, error) {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	var result []JobStatus
	for _, jobStatus := range repo.jobStatuses {
		if jobStatus.JobId == jobId {
			result = append(result, jobStatus)
		}
	}
	return result, nil
}

// GetByJobIdBusinessDate retrieves JobStatus structs for a specific job id and business date.
//
// Mutates receiver: yes (mutex)
func (repo *memoryRepo) GetByJobIdBusinessDate(jobId JobIdType, busDt common.Date) ([]JobStatus, error) {

	repo.mut.Lock()
	defer repo.mut.Unlock()

	var result []JobStatus
	for _, jobStatus := range repo.jobStatuses {
		if jobStatus.JobId == jobId && jobStatus.BusinessDate == busDt {
			result = append(result, jobStatus)
		}
	}
	return result, nil
}
