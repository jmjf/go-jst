package repo

import (
	"sync"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
)

type memoryRepo struct {
	jobStatuses []jobStatus.JobStatus
	mut         sync.Mutex
}

func NewMemoryRepo(jobStatuses []jobStatus.JobStatus) *memoryRepo {
	return &memoryRepo{jobStatuses: jobStatuses}
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: yes (mutex, data)
func (repo *memoryRepo) add(jobStatus jobStatus.JobStatus) error {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	repo.jobStatuses = append(repo.jobStatuses, jobStatus)
	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: yes (mutex)
func (repo *memoryRepo) GetByJobId(jobId jobStatus.JobIdType) ([]jobStatus.JobStatus, error) {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	var result []jobStatus.JobStatus
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
func (repo *memoryRepo) GetByJobIdBusinessDate(jobId jobStatus.JobIdType, busDt internal.Date) ([]jobStatus.JobStatus, error) {

	repo.mut.Lock()
	defer repo.mut.Unlock()

	var result []jobStatus.JobStatus
	for _, jobStatus := range repo.jobStatuses {
		if jobStatus.JobId == jobId && jobStatus.BusinessDate == busDt {
			result = append(result, jobStatus)
		}
	}
	return result, nil
}
