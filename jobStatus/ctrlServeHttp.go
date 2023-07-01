package jobStatus

import (
	"common"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type JobStatusCtrl interface {
	AddJobStatus(response http.ResponseWriter, request *http.Request, logger *slog.Logger)
}

type jobStatusCtrl struct {
	useCase JobStatusUC
}

// NewJobStatusCtrl creates and returns a JobStatusCtrl
func NewJobStatusCtrl(uc JobStatusUC) JobStatusCtrl {
	return &jobStatusCtrl{
		useCase: uc,
	}
}

// JobStatusCtrl.AddJobStatus attempts to add a new job status record to the database.
// If the request is invalid or adding fails, it logs errors and responds with
// an appropriate HTTP status code.
//
// Mutates receiver: no
func (jsc jobStatusCtrl) AddJobStatus(response http.ResponseWriter, request *http.Request, logger *slog.Logger) {

	// decode JSON into request data
	decoder := json.NewDecoder(request.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		//
		logErr := common.NewCommonError(err, common.ErrcdJsonDecode, request.Body)
		common.LogError(logger, "JSON Decode Error", logErr.Error(), logErr)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Debug("Call Add", "functionName", "jobStatusCtrl.AddJobStatus", "dto", dto)

	// call use case with DTO
	result, err := jsc.useCase.Add(dto)
	if err != nil {
		logErr := common.WrapError(err)
		// Need to identify error type and get it for logging
		var ce *common.CommonError
		var responseStatus int

		if errors.As(err, &ce) {
			responseStatus = http.StatusBadRequest
			common.LogError(logger, ce.Err.Error(), logErr.Error(), ce)
		} else {
			responseStatus = http.StatusInternalServerError
			logger.Error("Unknown error type", "err", err)
		}

		response.WriteHeader(responseStatus)
		return
	}

	logger.Debug("Add Result", "functionName", "jobStatusCtrl.AddJobStatus", "jobStatus", result)

	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(response)
	encoder.Encode(result)
}
