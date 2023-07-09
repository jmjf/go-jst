package repo

import (
	"sync"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
)

type repoDB struct {
	jobStatuses []jobStatus.JobStatus
	mut         sync.Mutex
}

func NewRepoDb(jobStatuses []jobStatus.JobStatus) *repoDB {
	return &repoDB{jobStatuses: jobStatuses}
}

// Open is a no-op for dbInMem
//
// Mutates receiver: no
func (repo *repoDB) Open() error {
	// in memory has nothing to open
	return nil
}

// Close is a no-op for dbInMem
//
// Mutates receiver: no
func (repo *repoDB) Close() error {
	// in memory has nothing to open
	return nil
}

// add inserts a JobStatus into the database.
//
// Mutates receiver: yes (mutex, data)
func (repo *repoDB) add(jobStatus jobStatus.JobStatus) error {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	repo.jobStatuses = append(repo.jobStatuses, jobStatus)
	return nil
}

// GetByJobId retrieves JobStatus structs for a specific job id.
//
// Mutates receiver: yes (mutex)
func (repo *repoDB) GetByJobId(jobId jobStatus.JobIdType) ([]jobStatus.JobStatus, error) {
	repo.mut.Lock()
	defer repo.mut.Unlock()

	var result []jobStatus.JobStatus
	for _, js := range repo.jobStatuses {
		if js.JobId == jobId {
			result = append(result, js)
		}
	}
	return result, nil
}

// GetByJobIdBusinessDate retrieves JobStatus structs for a specific job id and business date.
//
// Mutates receiver: yes (mutex)
func (repo *repoDB) GetByJobIdBusinessDate(jobId jobStatus.JobIdType, busDt internal.Date) ([]jobStatus.JobStatus, error) {

	repo.mut.Lock()
	defer repo.mut.Unlock()

	var result []jobStatus.JobStatus
	for _, js := range repo.jobStatuses {
		if js.JobId == jobId && js.BusinessDate == busDt {
			result = append(result, js)
		}
	}
	return result, nil
}
