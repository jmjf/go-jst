package jobStatus

import (
	"sync"
	"time"
)

type memoryRepo struct {
	jobStatuses []JobStatus
	mut         sync.Mutex
}

func NewMemoryRepo(jobStatuses []JobStatus) JobStatusRepo {
	return &memoryRepo{jobStatuses: jobStatuses}
}

// add inserts a JobStatus into the database.
func (repo *memoryRepo) Add(jobStatus JobStatus) error {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	repo.jobStatuses = append(repo.jobStatuses, jobStatus)
	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
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
func (repo *memoryRepo) GetByJobIdBusinessDate(jobId JobIdType, busDt time.Time) ([]JobStatus, error) {

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
