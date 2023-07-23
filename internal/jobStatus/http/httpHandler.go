package http

import (
	"errors"
	"log/slog"
	"net/http"

	"go-slo/internal"
	"go-slo/internal/jobStatus"
	"go-slo/internal/middleware"
)

func Handler(rootLogger *slog.Logger, ctrl *jobStatus.Controllers) http.Handler {
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
			ctrl.GetByQuery(res, req, logger)
			return

		case req.Method == http.MethodPost:
			ctrl.Add(res, req, logger)
			return

		default:
			logger.Error("Not Implemented")
			res.WriteHeader(http.StatusNotImplemented)
			return
		}
		// error checks, logging, etc.
	})
}
