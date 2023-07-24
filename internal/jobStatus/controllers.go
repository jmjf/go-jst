package jobStatus

import (
	"errors"
	"net/http"

	"go-slo/internal"
)

// CallUseCase calls a use case function with data and interprets the results.
// UCRT is the use case function's primary return type (secondary type is error).
// Do not use a pointer value for UCRT.
// DT is the type of data passed to the use case function.
//
// On success, CallUseCase returns a *UCRT. The caller must decide how to
// encode the results (data, statuses, etc.).
//
// This function is a controller (type of adapter) in the sense that it works for
// HTTP, gRPC or any other protocol. It uses HTTP status codes to communicate
// whether the result is ok, due to a problem with the request or due to an
// application or infrastructure error. The caller can decide how to translate the
// statuses for the protocol it supports.

// The caller is also responsible for translating inbound data to an expected format.
// For example, the caller must translate a JSON body, protobuf, etc., into a form the
// use case can use. It is also responsible for translating the results into a form
// the protocol can transport. So, CallUseCase is the business half of the controller,
// closer to the use cases,  and the caller is the infrastructure half, closer to the
// specific data transfer protocol.
func CallUseCase[UCRT any, DT any](ucFn func(DT) (UCRT, error), fnData DT) (*UCRT, int, error) {
	resStatus := http.StatusOK

	result, err := ucFn(fnData)
	if err != nil {
		resStatus = http.StatusInternalServerError // default error assumption

		var le *internal.LoggableError
		if errors.As(err, &le) {
			// if the error is caused by the request, set StatusBadRequest
			for _, cd := range internal.BadRequestErrcds {
				if le.Code == cd {
					resStatus = http.StatusBadRequest
					break
				}
			}
		}
	}

	return &result, resStatus, err
}
