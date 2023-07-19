package middleware

import (
	"context"
	"net/http"
	"strconv"

	"go-slo/internal"

	"github.com/jaevor/go-nanoid"
)

// requestIdKey is the reqeust id's name in the context.
const requestIdKey = "requestId"
const traceIdHeader = "X-Goslo-Trace-Id"

// GetRequestId accepts a context and attempts to get the request id from it.
// If it can get the request id, it returns it.
// If it cannot get the request id, it returns 0.
// TODO: return an error if cannot get request id
func GetRequestId(ctx context.Context) (string, error) {
	reqId, ok := ctx.Value(requestIdKey).(string)
	if !ok {
		return "", internal.NewLoggableError(internal.ErrMWGetReqId, internal.ErrcdMWGetReqId, reqId)
	}
	return reqId, nil
}

// AddRequestId returns a middleware handler that assigns a request id to the request's context.
func AddRequestId(next http.Handler) http.Handler {
	var reqId uint64
	useNano := true
	generateNano, err := nanoid.Standard(21)
	if err != nil {
		useNano = false
		reqId = uint64(1)
		// TODO: log the error
		// Low risk because, other than stdlib errors, the only error is invalid length--21 is valid.
	}

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var ctx context.Context

		traceId := req.Header.Get(traceIdHeader)
		if traceId == "" {
			var tracePrefix = req.URL.Path + "|" + req.Method + "|"
			if useNano {
				traceId = tracePrefix + generateNano()
			} else {
				traceId = tracePrefix + strconv.FormatUint(reqId, 36)
				reqId++
			}
		}
		ctx = context.WithValue(req.Context(), requestIdKey, traceId)
		res.Header().Set(traceIdHeader, traceId)

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}
