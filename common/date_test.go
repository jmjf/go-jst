package common_test

import (
	"common"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type TestStruct struct {
	Name string      `json:"name"`
	Dt   common.Date `json:"dt"`
}

const testDateStr = "2023-06-20"

var testDate, _ = common.NewDate(testDateStr)

func Test_date_String_ConvertsToString(t *testing.T) {
	expected := testDateStr
	dt, err := common.NewDate(expected)
	if err != nil {
		t.Errorf("FAIL | err %v", err)
		return
	}

	got := dt.String()

	if got != expected {
		t.Errorf("FAIL | expected %s | got %s", expected, got)
	}
}

func Test_date_Decode_ConvertsJsonToDate(t *testing.T) {
	expected := TestStruct{Name: "Joe", Dt: testDate}

	dec := json.NewDecoder((strings.NewReader(fmt.Sprintf(`{ "name": "Joe", "dt": "%s" }`, testDateStr))))
	var got TestStruct
	err := dec.Decode(&got)
	if err != nil {
		t.Errorf("FAIL | error json decoding %v | got %v", err, got)
		return
	}

	if expected.Dt != got.Dt || expected.Name != got.Name {
		t.Errorf("FAIL | expected %v | got %v", expected, got)
	}
}

func Test_date_Marshal_ConvertsDateToJson(t *testing.T) {
	expected := fmt.Sprintf(`{"name":"Joe","dt":"%s"}`, testDateStr)
	testStruct := TestStruct{Name: "Joe", Dt: testDate}

	got, err := json.Marshal(testStruct)
	if err != nil {
		t.Errorf("FAIL | error %v | got %s", err, string(got))
		return
	}
	if string(got) != expected {
		t.Errorf("FAIL | expected %s | got %s", expected, string(got))
	}
}
