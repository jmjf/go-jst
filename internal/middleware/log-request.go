package middleware

import (
	"errors"
	"go-slo/internal"
	"log/slog"
	"math"
	"net/http"
	"time"
)

// resWrapper wraps http.ResponseWriter to get status and content length data about the response.
// This wrapper could be better (based on httpsnoop), but is okay for now. I expect to replace
// much of this code when I switch from net/http to a fancier router/mux package.
type resWrapper struct {
	http.ResponseWriter
	status        int
	contentLength int
}

func (rw *resWrapper) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *resWrapper) Write(data []byte) (int, error) {
	rw.contentLength += len(data)
	return rw.ResponseWriter.Write(data)
}

// wrapResponseWriter wraps the http.ResponseWriter in resWrapper so we can use the added features.
func wrapResponseWriter(res http.ResponseWriter) *resWrapper {
	return &resWrapper{ResponseWriter: res, contentLength: 0}
}

// requestLogLevel controls the log level used to log requests and make controlling logging volume easier.
const requestLogLevel = slog.LevelInfo

// BuildReqLoggerMw returns a function that can be used in a standard net/http middleware chain.
// This function exists because the request logger needs a logger. I want to keep the middleware handler
// that accepts the next handler consistent with normal patterns.
func BuildReqLoggerMw(logger *slog.Logger) func(http.Handler) http.Handler {
	// logger is available in this closure space, so we can use it below safely.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			// move on if the logger won't log the target level
			if !logger.Handler().Enabled(nil, requestLogLevel) {
				next.ServeHTTP(res, req)
				return
			}

			rcvTs := time.Now().UTC()
			requestId, err := GetRequestId(req.Context())
			if err != nil {
				logErr := internal.WrapError(err)
				var le *internal.LoggableError
				if errors.As(err, &le) {
					internal.LogError(logger, le.Err.Error(), logErr.Error(), le)
				} else {
					logger.Error("Unknown error type", "err", err)
				}
				// not fatal: log but continue with 0 requestId
			}

			logger.Log(nil, requestLogLevel, "received", "remoteAddr", req.RemoteAddr,
				"requestId", requestId,
				"requestURI", req.RequestURI,
				"method", req.Method,
				"receivedContentLength", req.ContentLength,
				"receivedTime", rcvTs.Format(time.RFC3339Nano),
			)

			wrappedRes := wrapResponseWriter(res)

			defer func() {
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

				// TODO: Replace with a better solution.
				updateStats(wrappedRes, req, resTm)
				logger.Info("stats", "routeStats", GetRouteStats())
			}()

			next.ServeHTTP(wrappedRes, req)
		})
	}
}

// RouteStats holds basic request/response statistics for the server for monitoring purposes.
// TODO: Replace this code with a better solution.
type RouteStats struct {
	RequestCount   int
	ResponseCount  int
	TotalExecTime  time.Duration
	Status200Count int // 200-series statuses
	Status400Count int // 400-series statuses
	Status500Count int // 500-series statuses
}

// routeStatus holds statistics for each route called to date.
var routeStats = make(map[string]RouteStats)

// updateStats applies data related to a single request to the route statistics in routeStats.
func updateStats(wrappedRes *resWrapper, req *http.Request, resTm time.Duration) {
	statsKey := req.Method + "|" + req.RequestURI

	st, ok := routeStats[statsKey]
	if !ok {
		st = RouteStats{RequestCount: 1}
	} else {
		st.RequestCount++
	}
	st.TotalExecTime += resTm
	switch {
	case wrappedRes.status >= 200 && wrappedRes.status <= 299:
		st.Status200Count++
	case wrappedRes.status >= 400 && wrappedRes.status <= 499:
		st.Status400Count++
	case wrappedRes.status >= 500 && wrappedRes.status <= 599:
		st.Status500Count++
	}

	routeStats[statsKey] = st
}

func GetRouteStats() map[string]RouteStats {
	return routeStats
}
