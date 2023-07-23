package jobStatus

import (
	"errors"

	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

// UseCase is the set of use cases available for this subdomain and the major
// adapters (repos, external service, etc.) tand other reference objects they
// need to perform business processes against data.
type UseCase struct {
	repo Repo
}

// NewJobStatusUC creates and returns an UseCase“.
func NewUseCase(jsr Repo) *UseCase {
	return &UseCase{
		repo: jsr,
	}
}

// Add attempts to add a new job status to the data store.
// Returns a JobStatus and nil error on success.
// Returns an error (LoggableError) and empty JobStatus on failure.
//
// Mutates receiver: no
func (uc UseCase) Add(dto dtoType.JobStatusDto) (JobStatus, error) {
	jobStatus, err := NewJobStatus(dto)
	if err != nil {
		return JobStatus{}, internal.WrapError(err)
	}

	err = uc.repo.Add(jobStatus)
	if err != nil {
		return JobStatus{}, internal.WrapError(err)
	}

	return jobStatus, nil
}

// GetByQuery finds job status data for a query (string).
// Returns a slice of JobStatus and nil error on success.
// Returns an error (LoggableError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc UseCase) GetByQuery(rawQuery RequestQuery) ([]JobStatus, error) {
	if len(rawQuery) == 0 ||
		(len(rawQuery["jobId"]) == 0 && len(rawQuery["applicationId"]) == 0) ||
		(len(rawQuery["jobStatusTimestamp"]) == 0 && len(rawQuery["businessDate"]) == 0) {
		return []JobStatus{}, internal.NewLoggableError(internal.ErrAppTermMissing, internal.ErrcdAppTermMissing, rawQuery)
	}

	queryMap := make(map[string]string)
	for jt, fn := range validFields {
		tagVal := rawQuery.Get(jt)
		// Get() returns the first value only or "" if the array is empty.
		// Ignore net empty values because we can't query against them.
		if len(tagVal) == 0 {
			continue
		}
		queryMap[fn] = tagVal
	}

	result, err := uc.repo.GetByQuery(queryMap)
	if err != nil {
		// only error if error isn't NotFound
		var le *internal.LoggableError
		if errors.As(err, &le) {
			if le.Code != internal.ErrcdRepoNotFound {
				return []JobStatus{}, internal.WrapError(err)
			}
		}
		// NotFound, ensure result is empty
		result = []JobStatus{}
	}

	return result, nil
}
