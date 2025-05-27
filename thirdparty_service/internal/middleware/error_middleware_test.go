package middleware

import (
	"encoding/json"
	"errors"
	"testing"
	"thirdparty_service/internal/dtos"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp" // Import fasthttp
)

// Helper function to create a mock Fiber context
func createMockContext() *fiber.Ctx {
	app := fiber.New()
	// Acquire a new context and its underlying fasthttp context
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	fCtx := ctx.Context() // Corrected method to access fasthttp context

	// Create and set the fasthttp request and response
	fCtx.Request = fasthttp.Request{}
	fCtx.Response = fasthttp.Response{}

	return ctx
}

func TestFiberErrorHandler_ResponseError(t *testing.T) {
	ctx := createMockContext()
	defer ctx.App().ReleaseCtx(ctx) // Release the context

	expectedResponse := dtos.Response{
		Code: fiber.StatusBadRequest,
		Msg:  "Bad Request",
	}
	err := expectedResponse // Use the dtos.Response as the error

	handlerErr := FiberErrorHandler(ctx, err)

	assert.Nil(t, handlerErr, "Handler should not return an error")
	assert.Equal(t, expectedResponse.Code, ctx.Response().StatusCode(), "Status code should match the response code")

	var actualResponse dtos.Response
	unmarshalErr := json.Unmarshal(ctx.Response().Body(), &actualResponse) // Read from the response body
	assert.Nil(t, unmarshalErr, "Should not have error unmarshalling response body")
	assert.Equal(t, expectedResponse, actualResponse, "Response body should match the expected response")
}

func TestFiberErrorHandler_ErrorError(t *testing.T) {
	ctx := createMockContext()
	defer ctx.App().ReleaseCtx(ctx) // Release the context

	expectedError := dtos.Error{
		Code: fiber.StatusNotFound,
		Err:  "Not Found",
	}
	err := expectedError // Use the dtos.Error as the error

	handlerErr := FiberErrorHandler(ctx, err)

	assert.Nil(t, handlerErr, "Handler should not return an error")
	assert.Equal(t, expectedError.Code, ctx.Response().StatusCode(), "Status code should match the error code")

	var actualError dtos.Error
	unmarshalErr := json.Unmarshal(ctx.Response().Body(), &actualError) // Read from the response body
	assert.Nil(t, unmarshalErr, "Should not have error unmarshalling response body")
	assert.Equal(t, expectedError, actualError, "Response body should match the expected error")
}

func TestFiberErrorHandler_GenericError(t *testing.T) {
	ctx := createMockContext()
	defer ctx.App().ReleaseCtx(ctx) // Release the context

	genericErr := errors.New("something went wrong")

	handlerErr := FiberErrorHandler(ctx, genericErr)

	assert.Nil(t, handlerErr, "Handler should not return an error")
	assert.Equal(t, fiber.StatusInternalServerError, ctx.Response().StatusCode(), "Status code should be 500 for generic errors")

	var actualError dtos.Error
	unmarshalErr := json.Unmarshal(ctx.Response().Body(), &actualError) // Read from the response body
	assert.Nil(t, unmarshalErr, "Should not have error unmarshalling response body")
	assert.Equal(t, fiber.StatusInternalServerError, actualError.Code, "Error code should be 500")
	assert.Equal(t, "Internal Server Error", actualError.Err, "Error message should be 'Internal Server Error'")
}