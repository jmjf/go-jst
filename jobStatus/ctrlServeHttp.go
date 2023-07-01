package jobStatus

import (
	"common"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type JobStatusCtrl interface {
	AddJobStatus(response http.ResponseWriter, request *http.Request, logger *slog.Logger)
}

type jobStatusCtrl struct {
	useCase JobStatusUC
}

func NewJobStatusController(uc JobStatusUC) JobStatusCtrl {
	return &jobStatusCtrl{
		useCase: uc,
	}
}

func (jsc jobStatusCtrl) AddJobStatus(response http.ResponseWriter, request *http.Request, logger *slog.Logger) {

	// decode JSON into request data
	decoder := json.NewDecoder(request.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		logErr := common.NewCtrlError(err, common.ErrcdJsonDecode, request.Body)
		logger.Error("JSON Decode Error", "err", logErr) // TODO improve error formatting
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Debug("Call Add", "functionName", "jobStatusCtrl.AddJobStatus", "dto", dto)

	// call use case with DTO
	result, err := jsc.useCase.Add(dto)
	if err != nil {
		logErr := common.WrapError(err)
		// Need to identify error type and get it for logging
		var domainErr *common.DomainError
		var repoErr *common.RepoError
		var appErr *common.AppError
		var baseErr *common.BaseError
		var responseStatus int

		if errors.As(err, &domainErr) {
			responseStatus = http.StatusBadRequest
			logger.Error(domainErr.Err.Error(),
				slog.String("callStack", logErr.Error()),
				slog.String("fileName", repoErr.FileName),
				slog.String("funcName", repoErr.FuncName),
				slog.Int("lineNo", repoErr.LineNo),
				"errorData", fmt.Sprintf("%+v", domainErr.Data),
			)
		} else if errors.As(err, &repoErr) {
			switch repoErr.Code {
			case common.ErrcdRepoDupeRow:
				responseStatus = http.StatusConflict
			default:
				responseStatus = http.StatusInternalServerError
			}

			logger.Error(repoErr.Err.Error(),
				slog.String("callStack", logErr.Error()),
				slog.String("fileName", repoErr.FileName),
				slog.String("funcName", repoErr.FuncName),
				slog.Int("lineNo", repoErr.LineNo),
				slog.String("code", repoErr.Code),
			)
		} else if errors.As(err, &appErr) {
			responseStatus = http.StatusInternalServerError
			logger.Error(appErr.Err.Error(),
				slog.String("callStack", logErr.Error()),
				slog.String("fileName", appErr.FileName),
				slog.String("funcName", appErr.FuncName),
				slog.Int("lineNo", appErr.LineNo),
				slog.String("code", appErr.Code),
			)
		} else if errors.As(err, &baseErr) {
			responseStatus = http.StatusInternalServerError
			logger.Error("UnknownError", "callStack", logErr, "errorData", baseErr)
		} else {
			responseStatus = http.StatusInternalServerError
			logger.Error("Unknown error type", "err", err)
		}

		response.WriteHeader(responseStatus)
		return
	}

	logger.Debug("Add Result", "functionName", "jobStatusCtrl.AddJobStatus", "jobStatus", result)

	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(response)
	encoder.Encode(result)
}
