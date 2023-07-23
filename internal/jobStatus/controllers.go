package jobStatus

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"runtime"

	"go-slo/internal"
	dtoType "go-slo/public/jobStatus/http/20230701"
)

type RequestQuery = url.Values
type reqData[T any] struct {
	val T
}

type Controllers struct {
	uc  *UseCases
	log *slog.Logger
}

// NewControllers creates and returns a Controllers
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
func (ctrl Controllers) Add(res http.ResponseWriter, req *http.Request) {

	// decode JSON into request data
	decoder := json.NewDecoder(req.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto dtoType.JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		logErr := internal.NewLoggableError(err, internal.ErrcdJsonDecode, req.Body)
		internal.LogError(ctrl.log, "JSONDecodeError", logErr.Error(), logErr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// call use case with DTO
	result, resStatus, err := callUseCase[JobStatus, dtoType.JobStatusDto](ctrl.log, ctrl.uc.Add, dto)
	if err != nil {
		res.WriteHeader(resStatus)
		return
	}

	res.WriteHeader(http.StatusOK)
	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)
}

// GetByQuery attempts to find job statuses in the database that match a query string.
// If the query string is invalid or the query fails, it logs errors and responds with an appropriate HTTP status code.
//
// Mutates receiver: no
func (ctrl Controllers) GetByQuery(res http.ResponseWriter, req *http.Request) {

	// call use case with query
	result, resStatus, err := callUseCase[[]JobStatus, RequestQuery](ctrl.log, ctrl.uc.GetByQuery, req.URL.Query())
	if err != nil {
		res.WriteHeader(resStatus)
		return
	}

	res.WriteHeader(http.StatusOK)
	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)
}

// callUseCase abstracts calling a use case function and handling the error.
// UCRT is the use case function's primary return type (secondary type is error).
// (Do not use a pointer value for UCRT)
// DT is the type of the data passed to the use case function.
//
// On success, callUseCase returns a *UCRT. The caller must decide how to
// encode the results (data, statuses, etc.).
func callUseCase[UCRT any, DT any](log *slog.Logger, ucFn func(DT) (UCRT, error), fnData DT) (*UCRT, int, error) {
	resStatus := http.StatusOK
	callerNm := "runtime.Caller(1) error"
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		callerNm = runtime.FuncForPC(pc).Name()
	}
	log.Info("callUseCase", "callerNm", callerNm, "fnData", fnData)

	result, err := ucFn(fnData)
	if err != nil {
		logErr := internal.WrapError(err)
		// Need to identify error type and get it for logging
		var le *internal.LoggableError

		// TODO: Get correct error status for loggable error.
		// The error may be an internal server error (e.g., database error).
		// I need a way to choose a response code based on the specific error.
		if errors.As(err, &le) {
			resStatus = http.StatusBadRequest
			internal.LogError(log, le.Err.Error(), logErr.Error(), le)
		} else {
			resStatus = http.StatusInternalServerError
			log.Error("Unknown error type", "err", err)
		}
	}

	log.Info("callUseCase result", "callerNm", callerNm, "result", result, "resStatus", resStatus, "err", err)

	return &result, resStatus, err
}
