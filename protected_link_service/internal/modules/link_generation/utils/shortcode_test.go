package link_utils_test

import (
	link_utils "protected_link/internal/modules/link_generation/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateShortCode_Length(t *testing.T) {
	code := link_utils.GenerateShortCode(8)
	assert.Len(t, code, 8)
}

func TestGenerateShortCode_Uniqueness(t *testing.T) {
	code1 := link_utils.GenerateShortCode(8)
	time.Sleep(time.Millisecond) // ensure different seed
	code2 := link_utils.GenerateShortCode(8)
	assert.NotEqual(t, code1, code2)
}

func TestConvertToGenerateUrlRequest_Success(t *testing.T) {
	input := map[string]interface{}{
		"user_id":      "u1",
		"name":         "n1",
		"request_type": "auth",
		"model_type":   "jwt",
		"email":        "e@e.com",
		"expire_in":    "1h",
		"otp_required": false,
		"phone":        "123",
		"channel_type": "email",
		"data":         map[string]interface{}{"foo": "bar"},
	}
	res, err := link_utils.ConvertToGenerateUrlRequest(input)
	assert.NoError(t, err)
	assert.Equal(t, "u1", res.UserID)
	assert.Equal(t, "bar", res.Data["foo"])
}

func TestConvertToGenerateUrlRequest_MarshalError(t *testing.T) {
	ch := make(chan int)
	_, err := link_utils.ConvertToGenerateUrlRequest(ch)
	assert.Error(t, err)
}

func TestConvertToGenerateUrlRequest_UnmarshalError(t *testing.T) {
	input := map[string]interface{}{"user_id": make(chan int)}
	_, err := link_utils.ConvertToGenerateUrlRequest(input)
	assert.Error(t, err)
}

func TestParseExpiration_Success(t *testing.T) {
	ts, err := link_utils.ParseExpiration("1h")
	assert.NoError(t, err)
	assert.Greater(t, ts, time.Now().Unix())
}

func TestParseExpiration_Error(t *testing.T) {
	_, err := link_utils.ParseExpiration("notaduration")
	assert.Error(t, err)
}
