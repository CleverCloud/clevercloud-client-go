package client

import "net/http"

// Authenticator is called to authenticate an HTTP request
type Authenticator interface {
	// Add authentication stuff on an HTTP request
	Sign(*http.Request)
}
