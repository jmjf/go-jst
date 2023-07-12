package middleware

import (
	"context"
	"net/http"
)

// requestIdKey is the reqeust id's name in the context.
const requestIdKey = "requestId"

// GetRequestId accepts a context and attempts to get the request id from it.
// If it can get the request id, it returns it.
// If it cannot get the request id, it returns 0.
// TODO: return an error if cannot get request id
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
