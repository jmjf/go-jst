package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// requestLogLevel controls the log level used to log requests and make controlling logging volume easier.
const requestLogLevel = slog.LevelInfo

// LogRequest logs information about received requests and their responses.
func LogRequest(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// move on if the logger won't log the target level
		if !logger.Handler().Enabled(nil, requestLogLevel) {
			next.ServeHTTP(res, req)
			return
		}

		rcvTs := time.Now().UTC()
		requestId := GetRequestId(req.Context())

		logger.Log(nil, requestLogLevel, "received", "remoteAddr", req.RemoteAddr,
			"requestId", requestId,
			"requestURI", req.RequestURI,
			"method", req.Method,
			"receivedContentLength", req.ContentLength,
			"receivedTime", rcvTs.Format(time.RFC3339Nano),
		)

		next.ServeHTTP(res, req)

		resTm := time.Since(rcvTs)
		logger.Log(nil, requestLogLevel, "responding", "remoteAddr", req.RemoteAddr,
			"requestId", requestId,
			"requestURI", req.RequestURI,
			"method", req.Method,
			"receivedContentLength", req.ContentLength,
			"receivedTime", rcvTs.Format(time.RFC3339Nano),
			"responseMs", float64(resTm.Nanoseconds())/1000000.0,
			// "statusCode", req.Response.StatusCode,
			// "responseContentLength", req.Response.ContentLength,
		)
	})
}
