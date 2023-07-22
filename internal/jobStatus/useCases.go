package jobStatus

import (
	"errors"
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
// Returns an error (LoggableError) and empty JobStatus on failure.
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

type GetByQueryUC struct {
	jobStatusRepo Repo
}

// NewGetByQueryUC creates and returns a GetByQueryUC
func NewGetByQueryUC(jsr Repo) *GetByQueryUC {
	return &GetByQueryUC{
		jobStatusRepo: jsr,
	}
}

// TODO: GetByQueryUC.Execute finds job status data for a query (string).
// Returns a slice of JobStatus and nil error on success.
// Returns an error (LoggableError) and empty slice of JobStatus on failure.
//
// Mutates receiver: no
func (uc GetByQueryUC) Execute(rawQuery RequestQuery) ([]JobStatus, error) {
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

	result, err := uc.jobStatusRepo.GetByQuery(queryMap)
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
