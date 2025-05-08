package utils

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSendSuccessResponse(t *testing.T) {
	// Basic Success Response
	resp := SendSuccessResponse(200, "OK", map[string]string{"foo": "bar"})
	if resp.Status != "success" || resp.StatusCode != 200 || resp.Message != "OK" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if data, ok := resp.Data.(map[string]string); !ok || data["foo"] != "bar" {
		t.Errorf("unexpected data: %+v", resp.Data)
	}

	// Nil Data
	resp = SendSuccessResponse(201, "Created", nil)
	if resp.Data != nil {
		t.Errorf("expected nil data, got: %+v", resp.Data)
	}

	// Empty Message
	resp = SendSuccessResponse(202, "", "test")
	if resp.Message != "" {
		t.Errorf("expected empty message, got: %s", resp.Message)
	}

	// Zero Status Code
	resp = SendSuccessResponse(0, "Zero", 123)
	if resp.StatusCode != 0 {
		t.Errorf("expected status code 0, got: %d", resp.StatusCode)
	}
}

func TestSendErrorResponse(t *testing.T) {
	// Basic Error Response
	errResp := SendErrorResponse(400, "Bad Request", "invalid input")
	if errResp.Status != "error" || errResp.StatusCode != 400 || errResp.Message != "Bad Request" {
		t.Errorf("unexpected error response: %+v", errResp)
	}
	if errResp.Error != "invalid input" {
		t.Errorf("unexpected error: %+v", errResp.Error)
	}

	// Nil Error
	errResp = SendErrorResponse(404, "Not Found", nil)
	if errResp.Error != nil {
		t.Errorf("expected nil error, got: %+v", errResp.Error)
	}

	// Empty Message
	errResp = SendErrorResponse(500, "", "server error")
	if errResp.Message != "" {
		t.Errorf("expected empty message, got: %s", errResp.Message)
	}

	// Zero Status Code
	errResp = SendErrorResponse(0, "Zero", 123)
	if errResp.StatusCode != 0 {
		t.Errorf("expected status code 0, got: %d", errResp.StatusCode)
	}
}

func TestStructInitializationAndJSONTags(t *testing.T) {
	// ValidationError
	vErr := ValidationError{Field: "email", Message: "invalid"}
	if vErr.Field != "email" || vErr.Message != "invalid" {
		t.Errorf("unexpected ValidationError: %+v", vErr)
	}
	b, err := json.Marshal(vErr)
	if err != nil {
		t.Errorf("json marshal failed: %v", err)
	}
	var m map[string]string
	json.Unmarshal(b, &m)
	if m["field"] != "email" || m["message"] != "invalid" {
		t.Errorf("unexpected json: %v", m)
	}

	// SuccessResponse
	sResp := SuccessResponse{"success", 200, "ok", "data"}
	b, err = json.Marshal(sResp)
	if err != nil {
		t.Errorf("json marshal failed: %v", err)
	}
	var m2 map[string]interface{}
	json.Unmarshal(b, &m2)
	if m2["status"] != "success" || int(m2["status_code"].(float64)) != 200 || m2["message"] != "ok" || m2["data"] != "data" {
		t.Errorf("unexpected json: %v", m2)
	}

	// ErrorResponse
	eResp := ErrorResponse{"error", 400, "fail", "err"}
	b, err = json.Marshal(eResp)
	if err != nil {
		t.Errorf("json marshal failed: %v", err)
	}
	var m3 map[string]interface{}
	json.Unmarshal(b, &m3)
	if m3["status"] != "error" || int(m3["status_code"].(float64)) != 400 || m3["message"] != "fail" || m3["error"] != "err" {
		t.Errorf("unexpected json: %v", m3)
	}

	// Check struct field tags via reflection
	typeTag := reflect.TypeOf(ValidationError{})
	if typeTag.Field(0).Tag.Get("json") != "field" || typeTag.Field(1).Tag.Get("json") != "message" {
		t.Errorf("unexpected json tags: %v", typeTag)
	}
}
