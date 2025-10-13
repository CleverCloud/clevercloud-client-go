package client

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFromHTTPResponse_PlainText(t *testing.T) {
	// Arrange
	plainTextBody := "This is a plain text response"
	httpRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(plainTextBody)),
		Header:     make(http.Header),
	}
	httpRes.Header.Set("Content-Type", "text/plain")

	// Act
	res := fromHTTPResponse[PlainTextString](httpRes)

	// Assert
	if res.HasError() {
		t.Errorf("Expected no error, got: %v", res.Error())
	}

	if res.StatusCode() != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode())
	}

	payload := res.Payload()
	if payload == nil {
		t.Fatal("Expected payload to be non-nil")
	}

	expectedPayload := PlainTextString(plainTextBody)
	actualPayload := string(*payload)
	expectedString := string(expectedPayload)
	if actualPayload != expectedString {
		t.Errorf("Expected payload %q, got %q", expectedString, actualPayload)
	}
}

func TestFromHTTPResponse_JSON(t *testing.T) {
	// Arrange
	type TestStruct struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	jsonBody := `{"message":"success","code":200}`
	httpRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
		Header:     make(http.Header),
	}
	httpRes.Header.Set("Content-Type", "application/json")

	// Act
	res := fromHTTPResponse[TestStruct](httpRes)

	// Assert
	if res.HasError() {
		t.Errorf("Expected no error, got: %v", res.Error())
	}

	payload := res.Payload()
	if payload == nil {
		t.Fatal("Expected payload to be non-nil")
	}

	if payload.Message != "success" {
		t.Errorf("Expected message 'success', got %q", payload.Message)
	}

	if payload.Code != 200 {
		t.Errorf("Expected code 200, got %d", payload.Code)
	}
}

func TestFromHTTPResponse_Error(t *testing.T) {
	// Arrange
	errorBody := "Internal Server Error"
	httpRes := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(errorBody)),
		Header:     make(http.Header),
	}

	// Act
	res := fromHTTPResponse[PlainTextString](httpRes)

	// Assert
	if !res.HasError() {
		t.Error("Expected an error")
	}

	if res.StatusCode() != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, res.StatusCode())
	}

	if res.Error() == nil {
		t.Error("Expected error to be non-nil")
	}
}
