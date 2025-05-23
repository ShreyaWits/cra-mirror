package dto

import (
	"encoding/json"
	"testing"
    "github.com/stretchr/testify/assert"
)

func TestDeleteTemplateRequest_JSONMarshaling(t *testing.T) {
	req := DeleteTemplateRequest{
		TemplateID: "template-123",
	}
	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"template_id":"template-123"}`, string(data))

	var unmarshaled DeleteTemplateRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, req, unmarshaled)
}

func TestDeleteTemplateResponse_JSONMarshaling(t *testing.T) {
	resp := DeleteTemplateResponse{
		Success: true,
		Message: "Template deleted successfully",
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"success":true,"message":"Template deleted successfully"}`, string(data))

	var unmarshaled DeleteTemplateResponse
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, resp, unmarshaled)
}
