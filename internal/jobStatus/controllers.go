package jobStatus

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

type RequestQuery = url.Values

type Controllers struct {
	uc  *UseCases
	log *slog.Logger
}

// NewControllers creates and returns an Controllers
func NewControllers(uc *UseCases, logger *slog.Logger) *Controllers {
	return &Controllers{
		uc:  uc,
		log: logger,
	}
}

// Add attempts to add a new job status record to the database.
// If the request is invalid or adding fails, it logs errors and responds with
// an appropriate HTTP status code.
//
// Mutates receiver: no
func (ctrl Controllers) Add(response http.ResponseWriter, request *http.Request) {

	// decode JSON into request data
	decoder := json.NewDecoder(request.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto dtoType.JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		//
		logErr := internal.NewLoggableError(err, internal.ErrcdJsonDecode, request.Body)
		internal.LogError(ctrl.log, "JSON Decode Error", logErr.Error(), logErr)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	ctrl.log.Debug("Call Execute", "functionName", "jobStatusCtrl.AddJobStatus", "dto", dto)

	// call use case with DTO
	result, err := ctrl.uc.Add(dto)
	if err != nil {
		logErr := internal.WrapError(err)
		// Need to identify error type and get it for logging
		var le *internal.LoggableError
		var responseStatus int

		if errors.As(err, &le) {
			responseStatus = http.StatusBadRequest
			internal.LogError(ctrl.log, le.Err.Error(), logErr.Error(), le)
		} else {
			responseStatus = http.StatusInternalServerError
			ctrl.log.Error("Unknown error type", "err", err)
		}

		response.WriteHeader(responseStatus)
		return
	}

	ctrl.log.Debug("Add Result", "functionName", "jobStatusCtrl.AddJobStatus", "jobStatus", result)

	response.WriteHeader(http.StatusOK)
	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(response)
	encoder.Encode(result)
}

// GetByQuery attempts to find job statuses in the database that match a query string.
// If the query string is invalid or the query fails, it logs errors and responds with an appropriate HTTP status code.
//
// Mutates receiver: no
func (ctrl Controllers) GetByQuery(res http.ResponseWriter, req *http.Request) {
	ctrl.log.Debug("Call Execute", "functionName", "jobStatusCtrl.AddJobStatus", "query", req.URL.Query())

	// call use case with DTO
	result, err := ctrl.uc.GetByQuery(req.URL.Query())
	if err != nil {
		logErr := internal.WrapError(err)
		// Need to identify error type and get it for logging
		var le *internal.LoggableError
		var resStatus int

		if errors.As(err, &le) {
			resStatus = http.StatusBadRequest
			internal.LogError(ctrl.log, le.Err.Error(), logErr.Error(), le)
		} else {
			resStatus = http.StatusInternalServerError
			ctrl.log.Error("Unknown error type", "err", err)
		}

		res.WriteHeader(resStatus)
		return
	}

	ctrl.log.Debug("Add Result", "functionName", "jobStatusCtrl.AddJobStatus", "jobStatus", result)

	res.WriteHeader(http.StatusOK)
	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)

}
