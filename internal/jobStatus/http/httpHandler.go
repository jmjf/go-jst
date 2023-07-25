package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	"go-slo/internal/middleware"

	dtoType "go-slo/public/jobStatus/http/20230701"
)

// Handler returns an http.Handler we can use with a mux.
// The returned handler will call the controller function that calls the use case for the
// request.
//
// The handler is a thin shim that uses request HTTP methods and request details to decide
// which controller function to run.
//
// Controller functions are responsible for decoding request bodies, calling the use case,
// setting HTTP response status, encoding response bodies, handling any errors, etc.
func Handler(rootLogger *slog.Logger, uc *jobStatus.UseCases) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		logger := rootLogger.With("route", req.URL.Path, "method", req.Method)
		requestId, err := middleware.GetRequestId(req.Context())
		if err != nil {
			logError(internal.WrapError(err), logger)
		}
		logger = logger.With("requestId", requestId)

		switch {
		// case req.Method == http.MethodGet && len(req.URL.Query()) == 0:
		// 	// handle get all
		case req.Method == http.MethodGet && len(req.URL.Query()) > 0:
			getByQueryCtrl(res, req, uc, logger)
			return
		case req.Method == http.MethodPost:
			addCtrl(res, req, uc, logger)
			return
		default:
			le := internal.NewLoggableError(internal.ErrNotImplemented, internal.ErrcdNotImplemented, "job status http handler")
			logError(le, logger)
			res.WriteHeader(http.StatusNotImplemented)
			return
		}
	})
}

// addCtrl uses the use case Add to attempt to add a new job status to the database.
// If the use case fails, it logs errors and sets the responseWriter with an appropriate HTTP status code.
func addCtrl(res http.ResponseWriter, req *http.Request, uc *jobStatus.UseCases, logger *slog.Logger) {
	// decode JSON into DTO
	// for now, we can convert the input to a DTO directly
	// TODO: add in-message API version checking; requires a DTO change
	var dto dtoType.JobStatusDto
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&dto)
	if err != nil {
		err = internal.NewLoggableError(err, internal.ErrcdJsonDecode, req.Body)
		logError(err, logger)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// call use case with DTO
	result, err := uc.Add(dto)
	if err != nil {
		logError(internal.WrapError(err), logger)
		res.WriteHeader(errToHTTPStatus(err, logger))
		return
	}

	// encode response; set response status explicitly
	res.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)
	return
}

// getByQueryCtrl calls the use case GetByQuery to attempts to find job statuses that match a query string.
// If the use case fails, it logs errors and sets the responseWriter with an appropriate HTTP status code.
func getByQueryCtrl(res http.ResponseWriter, req *http.Request, uc *jobStatus.UseCases, logger *slog.Logger) {
	// call use case with query
	result, err := uc.GetByQuery(req.URL.Query())
	if err != nil {
		logError(internal.WrapError(err), logger)
		res.WriteHeader(errToHTTPStatus(err, logger))
		return
	}

	// encode response; set response status explicitly
	res.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(res)
	encoder.Encode(result)
	return
}

// badRequestErrCds lists errors that should return 400 Bad Request
var badRequestErrcds = []string{
	internal.ErrcdDomainProps,
	internal.ErrcdAppTermInvalid,
	internal.ErrcdAppTermMissing,
	internal.ErrcdRepoDupeRow,
	internal.ErrcdRepoInvalidQuery,
	internal.ErrcdJsonDecode,
}

// errToHTTPStatus takes an error and chooses a matching HTTP status code.
func errToHTTPStatus(err error, logger *slog.Logger) int {
	if err == nil {
		// Log a warning with information about the caller so we can quit doing this.
		caller := "Unknown"
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("%s::%s::%d <- %w", filepath.Base(file), runtime.FuncForPC(pc).Name(), line, err)
		}
		logger.Warn(internal.ErrcdNilError, "in", internal.WrapError(err), "calledBy", caller)
		// StatusOK may be wrong if StatusAccepted or similar is more correct.
		// But it prevents unexpected behavior if called with a nil error.
		return http.StatusOK
	}

	// Assume the error will be a 500 Internal Server Error
	resStatus := http.StatusInternalServerError

	var le *internal.LoggableError
	if errors.As(err, &le) {
		// If the error is in the bad request list, change resStatus
		for _, cd := range badRequestErrcds {
			if le.Code == cd {
				resStatus = http.StatusBadRequest
				break
			}
		}
	}
	return resStatus
}

// logError logs an error as a LoggableError or as an unknown (raw) error.
// It calls internal.WrapError() so the logged error contains a full stack trace.
func logError(err error, logger *slog.Logger) {
	if err == nil {
		// Log a warning with information about the caller so we can find it.
		logErr := internal.WrapError(err)
		logger.Warn("LogErrorWithNilErr", "err", logErr)
	}

	var le *internal.LoggableError
	if errors.As(err, &le) {
		internal.LogError(logger, le.Code, le.Error(), le)
	} else {
		logger.Error(internal.ErrcdUnknown, "err", internal.WrapError(err))
	}
}
