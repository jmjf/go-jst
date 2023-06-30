package jobStatus

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JobStatusCtrl interface {
	AddJobStatus(response http.ResponseWriter, request *http.Request)
}

type jobStatusCtrl struct {
	useCase JobStatusUC
}

func NewJobStatusController(uc JobStatusUC) JobStatusCtrl {
	return &jobStatusCtrl{
		useCase: uc,
	}
}

func (jsc jobStatusCtrl) AddJobStatus(response http.ResponseWriter, request *http.Request) {

	// decode JSON into request data
	decoder := json.NewDecoder(request.Body)

	// TODO: add in-message API version checking; requires a raw JSON structure separate from DTO

	// for now, we can convert the input to a DTO directly
	var dto JobStatusDto

	err := decoder.Decode(&dto)
	if err != nil {
		fmt.Println("json decode error", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("Controller | Call Add with dto %+v\n", dto)

	// call use case with DTO
	result, err := jsc.useCase.Add(dto)
	if err != nil {
		fmt.Printf("Controller | Add returned error %v\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("Controller | Add returned result %+v\n", result)

	// encode response (generic to all HTTP controllers)
	encoder := json.NewEncoder(response)
	encoder.Encode(result)
}
