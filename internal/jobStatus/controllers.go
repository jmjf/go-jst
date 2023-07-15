package jobStatus

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

type AddJobStatusCtrl struct {
	useCase *AddJobStatusUC
}

// NewAddJobStatusCtrl creates and returns an AddJobStatusCtrl
func NewAddJobStatusCtrl(uc *AddJobStatusUC) *AddJobStatusCtrl {
	return &AddJobStatusCtrl{
		useCase: uc,
	}
}

// JobStatusCtrl.AddJobStatus attempts to add a new job status record to the database.
// If the request is invalid or adding fails, it logs errors and responds with
// an appropriate HTTP status code.
//
// Mutates receiver: no
func (ctrl AddJobStatusCtrl) Execute(response http.ResponseWriter, request *http.Request, logger *slog.Logger) {

	// decode JSON into request data
	decoder := json.NewDecoder(request.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto dtoType.JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		//
		logErr := internal.NewLoggableError(err, internal.ErrcdJsonDecode, request.Body)
		internal.LogError(logger, "JSON Decode Error", logErr.Error(), logErr)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Debug("Call Execute", "functionName", "jobStatusCtrl.AddJobStatus", "dto", dto)

	// call use case with DTO
	result, err := ctrl.useCase.Execute(dto)
	if err != nil {
		logErr := internal.WrapError(err)
		// Need to identify error type and get it for logging
		var le *internal.LoggableError
		var responseStatus int

		if errors.As(err, &le) {
			responseStatus = http.StatusBadRequest
			internal.LogError(logger, le.Err.Error(), logErr.Error(), le)
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
	response.WriteHeader(http.StatusOK)
}
