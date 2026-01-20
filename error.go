package client

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (e *APIError) Equal(e2 *APIError) bool {
	if (e == nil) != (e2 == nil) {
		return false
	}
	if e.RequestID != e2.RequestID || e.Code != e2.Code || e.Message != e2.Message {
		return false
	}
	// what about context ?

	return true
}

type CCApiError struct {
	ID      uint64         `json:"id"`
	Message string         `json:"message"`
	Type    string         `json:"type"`
	Fields  map[string]any `json:"fields,omitempty"`
}

func (e *CCApiError) Into() *APIError {
	err := &APIError{
		Code:    fmt.Sprintf("%d", e.ID),
		Message: e.Message,
		Context: map[string]any{"type": e.Type},
	}

	for name, value := range e.Fields {
		err.Context[name] = value
	}

	return err
}

// Best effor way to grab informations from payload
func APIErrorFrom(payload []byte) *APIError {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()

	var e APIError
	if err := dec.Decode(&e); err == nil {
		return &e
	}

	dec = json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()

	var ccapiErr CCApiError
	if err := dec.Decode(&ccapiErr); err == nil {
		return ccapiErr.Into()
	}

	return nil
}
