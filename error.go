package client

import (
	"bytes"
	"encoding/json"
)

type APIError struct {
	RequestID string         `json:"apiRequestId"`
	Code      string         `json:"code"`
	Context   map[string]any `json:"context"`
	Message   string         `json:"error"`
}

func (e APIError) String() string {
	return e.Message
}

func (e APIError) Error() string {
	return e.Message
}

func APIErrorFrom(payload []byte) *APIError {
	var e APIError
	err := json.
		NewDecoder(bytes.NewReader(payload)).
		Decode(&e)
	if err != nil {
		return nil
	}

	return &e
}
