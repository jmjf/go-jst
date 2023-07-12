package middleware

import (
	"context"
	"net/http"
)

const requestIdKey = "requestId"

func GetRequestId(ctx context.Context) uint64 {
	reqId, ok := ctx.Value(requestIdKey).(uint64)
	if !ok {
		return 0
	}
	return reqId
}

// AddRequestId returns a middleware handler that assigns a request id to the request's context.
func AddRequestId(next http.Handler) http.Handler {
	reqId := uint64(1)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), requestIdKey, reqId)
		reqId++
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}
