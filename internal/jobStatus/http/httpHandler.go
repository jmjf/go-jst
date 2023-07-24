package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	"go-slo/internal/middleware"

	dtoType "go-slo/public/jobStatus/http/20230701"
)

func Handler(rootLogger *slog.Logger, uc *jobStatus.UseCases) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		requestId, err := middleware.GetRequestId(req.Context())
		if err != nil {
			logErr := internal.WrapError(err)
			var le *internal.LoggableError
			if errors.As(err, &le) {
				internal.LogError(rootLogger, le.Err.Error(), logErr.Error(), le)
			} else {
				rootLogger.Error("Unknown error type", "err", err)
			}
			// not fatal: log but continue with 0 requestId
		}
		logger := rootLogger.With("route", req.URL.Path, "method", req.Method, "requestId", requestId)

		switch {
		// case req.Method == http.MethodGet && len(req.URL.Query()) == 0:
		// 	// handle get all

		case req.Method == http.MethodGet && len(req.URL.Query()) > 0:
			ctrlErr := GetByQueryCtrl(res, req, uc)
			if ctrlErr != nil {
				var le *internal.LoggableError
				logErr := internal.WrapError(ctrlErr)
				if errors.As(ctrlErr, &le) {
					internal.LogError(logger, le.Code, logErr.Error(), le)
				} else {
					logger.Error(internal.ErrUnknown.Error(), "err", logErr)
				}
			}
			return

		case req.Method == http.MethodPost:
			ctrlErr := AddCtrl(res, req, uc)
			if ctrlErr != nil {
				var le *internal.LoggableError
				logErr := internal.WrapError(ctrlErr)
				if errors.As(ctrlErr, &le) {
					internal.LogError(logger, le.Code, logErr.Error(), le)
				} else {
					logger.Error(internal.ErrUnknown.Error(), "err", logErr)
				}
			}
			return

		default:
			le := internal.NewLoggableError(internal.ErrNotImplemented, internal.ErrcdNotImplemented, nil)
			internal.LogError(logger, le.Code, le.Error(), le)
			res.WriteHeader(http.StatusNotImplemented)
			return
		}
		// error checks, logging, etc.
	})
}

// AddCtrl uses the use case Add to attempt to add a new job status to the database.
// If the request is invalid or adding fails, it logs errors and responds with
// an appropriate HTTP status code.
func AddCtrl(res http.ResponseWriter, req *http.Request, uc *jobStatus.UseCases) error {

	// decode JSON into request data
	decoder := json.NewDecoder(req.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto dtoType.JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		err = internal.NewLoggableError(err, internal.ErrcdJsonDecode, req.Body)
		res.WriteHeader(http.StatusBadRequest)
		return err
	}

	// call use case with DTO
	result, resStatus, ucErr := jobStatus.CallUseCase[jobStatus.JobStatus, dtoType.JobStatusDto](uc.Add, dto)
	if ucErr != nil {
		var le *internal.LoggableError
		if errors.As(ucErr, &le) {
			err = internal.NewLoggableError(ucErr, le.Code, dto)
		} else {
			err = internal.NewLoggableError(ucErr, internal.ErrcdUnknown, dto)
		}

		res.WriteHeader(resStatus)
		return err
	}

	res.WriteHeader(http.StatusOK)
	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)

	return nil
}

// GetByQueryCtrl calls the use case GetByQuery to attempts to find
// job statuses in the database that match a query string. If the
// query string is invalid or the query fails, it logs errors and
// responds with an appropriate HTTP status code.
func GetByQueryCtrl(res http.ResponseWriter, req *http.Request, uc *jobStatus.UseCases) error {
	// call use case with query
	result, resStatus, ucErr := jobStatus.CallUseCase[[]jobStatus.JobStatus, jobStatus.RequestQuery](uc.GetByQuery, req.URL.Query())
	if ucErr != nil {
		var le *internal.LoggableError
		var err error

		if errors.As(ucErr, &le) {
			err = internal.NewLoggableError(ucErr, le.Code, req.URL.Query())
		} else {
			err = internal.NewLoggableError(ucErr, internal.ErrcdUnknown, req.URL.Query())
		}

		res.WriteHeader(resStatus)
		return err
	}

	res.WriteHeader(http.StatusOK)
	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)

	return nil
}
