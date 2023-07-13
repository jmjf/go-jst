package middleware

import (
	"log/slog"
	"math"
	"net/http"
	"time"
)

// requestLogLevel controls the log level used to log requests and make controlling logging volume easier.
const requestLogLevel = slog.LevelInfo

type resWriter struct {
	http.ResponseWriter
	status        int
	contentLength int
}

func (rw *resWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *resWriter) Write(data []byte) (int, error) {
	rw.contentLength += len(data)
	return rw.ResponseWriter.Write(data)
}

func wrapResponseWriter(res http.ResponseWriter) *resWriter {
	return &resWriter{ResponseWriter: res, contentLength: 0}
}

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

		wrappedRes := wrapResponseWriter(res)

		next.ServeHTTP(wrappedRes, req)

		resTm := time.Since(rcvTs)
		logger.Log(nil, requestLogLevel, "responding", "remoteAddr", req.RemoteAddr,
			"requestId", requestId,
			"requestURI", req.RequestURI,
			"method", req.Method,
			"receivedContentLength", req.ContentLength,
			"receivedTime", rcvTs.Format(time.RFC3339Nano),
			"responseMs", math.Round(float64(resTm.Microseconds())/100.0)/10.0, // nnn.n ms
			"statusCode", wrappedRes.status,
			"responseContentLength", wrappedRes.contentLength,
		)
	})
}
