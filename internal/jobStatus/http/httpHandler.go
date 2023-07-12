package http

import (
	"log/slog"
	"net/http"

	"go-slo/internal/jobStatus"
	"go-slo/lib/middleware"
)

func Handler(rootLogger *slog.Logger, ctrl jobStatus.Controller) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		requestId := middleware.GetRequestId(req.Context())
		logger := rootLogger.With("route", req.URL.Path, "method", req.Method, "requestId", requestId)

		switch {
		// case req.Method == http.MethodGet && len(req.URL.Query()) == 0:
		// 	// handle get all

		// case req.Method == http.MethodGet && len(req.URL.Query()) > 0:
		// 	// handle get by query

		case req.Method == http.MethodPost:
			ctrl.Execute(res, req, logger)
			return

		default:
			logger.Error("Not Implemented")
			res.WriteHeader(http.StatusNotImplemented)
			return
		}
		// error checks, logging, etc.
	})
}
