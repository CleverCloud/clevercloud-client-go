package client_test

import (
	"testing"

	"go.clever-cloud.dev/client"
)

func TestAPIErrorFrom(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    *client.APIError
	}{{
		name:    "error",
		payload: []byte(`{"apiRequestId": "someid", "code": "some.error.code", "error": "an error", "context": {}}`),
		want: &client.APIError{
			RequestID: "someid",
			Code:      "some.error.code",
			Message:   "an error",
		},
	}, {
		name:    "ccapi err",
		payload: []byte(`{"id":4004,"message":"The provided id doesn't belong to any app","type":"error"}`),
		want: &client.APIError{
			Code:    "4004",
			Message: "The provided id doesn't belong to any app",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.APIErrorFrom(tt.payload)
			if !tt.want.Equal(got) {
				t.Errorf("APIErrorFrom() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
