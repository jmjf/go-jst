package jobStatus

import (
	"common"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Need to use uppercase field names so reflect can see them to use tags
type JobStatusDto struct {
	AppId string      `json:"applicationId"`
	JobId string      `json:"jobId"`
	JobSt string      `json:"jobStatusCode"`
	JobTs time.Time   `json:"jobStatusTimestamp"`
	BusDt common.Date `json:"businessDate"`
	RunId string      `json:"runId"`
	HstId string      `json:"hostId"`
}

// isUsable checks data on the DTO to ensure it can be used to create a JobStatus.
//
// It returns an array of errors describing all data problems.
// If len(errs) == 0, the DTO is good.
func (dto JobStatusDto) isUsable() []error {
	errs := []error{}
	now := time.Now()

	if len(dto.AppId) == 0 || len(dto.AppId) > 200 {
		errs = append(errs, fmt.Errorf("invalid ApplicationId |%s|", dto.AppId))
	}

	if len(dto.JobId) == 0 || len(dto.JobId) > 200 {
		errs = append(errs, fmt.Errorf("invalid JobId |%s|", dto.JobId))
	}

	if len(dto.JobSt) == 0 || dto.jobStatusCode() == JobStatus_INVALID {
		errs = append(errs, fmt.Errorf("invalid JobStatusCode |%s|", dto.JobSt))
	}

	// now < JobTs -> JobTs is in the future
	if now.Compare(dto.JobTs) == -1 {
		errs = append(errs, fmt.Errorf("invalid JobTimestamp |%s|", dto.JobTs.Format(time.RFC3339)))
	}

	// now < BusDt -> BusDt is in the future
	if now.Compare(time.Time(dto.BusDt)) == -1 {
		errs = append(errs, fmt.Errorf("invalid BusinessDate |%s|", dto.BusDt))
	}

	// if JobTs < BusDt -> error
	// need to think about this for TZ near international date line
	if dto.JobTs.Compare(time.Time(dto.BusDt)) == -1 {
		errs = append(errs, fmt.Errorf("JobTimestamp is less than BusinessDate |%s| |%s|", dto.JobTs.Format(time.RFC3339), dto.BusDt))
	}

	if len(dto.RunId) > 50 {
		errs = append(errs, fmt.Errorf("RunId is over 50 characters |%s|", dto.RunId))
	}

	if len(dto.HstId) > 150 {
		errs = append(errs, fmt.Errorf("HostId is over 150 characters |%s|", dto.HstId))
	}

	return errs
}

// normalizeTimes normalizes JobTs and BusDt to expected precision.
//
// Returns a JobStatusDto with JobTs truncated to seconds and converted to UTC
// and BusDt truncated to day as UTC (not converted to UTC).
func (dto *JobStatusDto) normalizeTimes() {
	dto.JobTs = dto.JobTs.Truncate(time.Second).UTC()

	// Date type's time components are already zero
	// yr, mo, dy := dto.BusDt.Date()
	// dto.BusDt = common.NewDate(fmt.Sprintf("time.Date(yr, mo, dy, 0, 0, 0, 0, time.UTC)
}

// jobStatusCode returns the job status code if dto.JobSt in the list of valid job status codes.
// If code is not a valid job status code, it returns JobStatus_INVALID.
func (dto *JobStatusDto) jobStatusCode() JobStatusCodeType {
	for _, jsc := range validJobStatusCodes {
		if string(jsc) == strings.ToUpper(dto.JobSt) {
			return jsc
		}
	}
	return JobStatus_INVALID
}

type JobStatusCtrl interface {
	AddJobStatus(response http.ResponseWriter, request *http.Request)
}

type jobStatusCtrl struct {
	useCase JobStatusUC
}

func NewJobStatusController(uc JobStatusUC) JobStatusCtrl {
	return &jobStatusCtrl{
		useCase: uc,
	}
}

func (jsc jobStatusCtrl) AddJobStatus(response http.ResponseWriter, request *http.Request) {

	// decode JSON into request data
	decoder := json.NewDecoder(request.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		fmt.Println("json decode error", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("Controller | Call Add with dto %+v\n", dto)

	// call use case with DTO
	result, err := jsc.useCase.Add(dto)
	if err != nil {
		fmt.Printf("Controller | Add returned error %v\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("Controller | Add returned result %+v\n", result)

	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(response)
	encoder.Encode(result)
}
