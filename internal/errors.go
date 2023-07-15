package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"runtime"
)

// LoggableError holds data we want from all errors to support logging
type LoggableError struct {
	FileName string
	FuncName string
	LineNo   int
	Data     any
	Code     string
	Err      error
}

// LoggableError.Error() returns a string representation of an error.
//
// Mutates receiver: no
func (ce *LoggableError) Error() string {
	return fmt.Sprintf("%s::%s::%d Code %s | %v", ce.FileName, ce.FuncName, ce.LineNo, ce.Code, ce.Err)
}

// LoggableError.Unwrap() returns the error contained within the error.
//
// Mutates receiver: no
func (ce *LoggableError) Unwrap() error {
	return ce.Err
}

// NewLoggableError creates a LoggableError. It uses runtime.Caller(1) to get information
// about the caller to include in the error structure, reducing call boilerplate.
func NewLoggableError(err error, code string, data any) *LoggableError {
	// get information about the function that called this one
	pc, file, line, ok := runtime.Caller(1)

	newErr := LoggableError{}
	newErr.Code = code
	newErr.Err = err
	newErr.Data = data
	if ok {
		newErr.FileName = filepath.Base(file)
		newErr.FuncName = runtime.FuncForPC(pc).Name()
		newErr.LineNo = line
	}
	return &newErr
}

// primitive errors and error codes for domain errors
var (
	ErrDomainProps   = errors.New("props error")
	ErrcdDomainProps = "PropsError"
)

// Primitive errors an error codes for application errors
var (
	ErrAppUnexpected   = errors.New("unexpected error")
	ErrcdAppUnexpected = "UnexpectedError"
)

// Primitive errors an error codes for repo errors
var (
	ErrRepoNoDsn           = errors.New("dsn is empty")
	ErrcdRepoNoDsn         = "NoDsn"
	ErrRepoScan            = errors.New("scan error")
	ErrcdRepoScan          = "ScanError"
	ErrRepoDupeRow         = errors.New("duplicate row error")
	ErrcdRepoDupeRow       = "DuplicateRowError"
	ErrRepoConnException   = errors.New("connection exception error")
	ErrcdRepoConnException = "ConnectionExceptionError"
	ErrRepoOther           = errors.New("other error")
	ErrcdRepoOther         = "RepoOtherError"
	// TODO: examine database error and classify it
	// should retry? etc.
)

var (
	ErrMWGetReqId   = errors.New("error getting request id")
	ErrcdMWGetReqId = "GetReqId"
)

// Primitive errors and error codes for controller errors
var (
	ErrJsonDecode   = errors.New("json decode error")
	ErrcdJsonDecode = "JsonDecodeError"
)

// WrapError returns an error that wraps the passed error with information
// about the WrapError caller. When bubbling up errors, this simplifies
// wrapping and ensures consistent lightweight stack traces.
func WrapError(err error) error {
	// get information about the function that called this one
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("unknown caller <- %w", err)
	}
	return fmt.Errorf("%s::%s::%d <- %w", filepath.Base(file), runtime.FuncForPC(pc).Name(), line, err)
}

// isEmptyJson detects if a string contains only empty JSON structures
var isEmptyJson = regexp.MustCompile(`^[\[\],{}]+$`).MatchString

// LogError logs an error message using, applying a common pattern.
func LogError(logger *slog.Logger, msg string, callStack string, ce *LoggableError) {
	errData := ce.Data

	// When ce.Data is an array of errors, json.Marshal() returns [{}].
	// If json.Marshal() returns no usable data, use Sprintf hoping for something usable.
	eJson, _ := json.Marshal(errData)
	if isEmptyJson(string(eJson)) {
		errData = fmt.Sprintf("%v", ce.Data)
	}
	// else, use errData from ce.Data (at top of func)

	logger.Error(msg,
		slog.String("callStack", callStack),
		slog.String("fileName", ce.FileName),
		slog.String("funcName", ce.FuncName),
		slog.Int("lineNo", ce.LineNo),
		slog.Any("errorData", errData),
	)
}
