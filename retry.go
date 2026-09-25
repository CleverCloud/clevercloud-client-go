package client

import (
	"context"
	"net/http"
	"time"
)

// ServerErrorRetryDelay is how long ServerErrorRetryablePolicy waits before
// allowing the next attempt.
const ServerErrorRetryDelay = 500 * time.Millisecond

// RetryPolicy decides whether a failed request is sent again.
//
// It receives the context of the call, the request that failed, the error the
// API answered with, and retries, the number of retries already performed for
// this call — 0 on the first failure. Returning true sends the request again.
//
// The policy owns its own backoff: the client sends the next request as soon as
// the policy returns, so a policy that wants to wait waits before returning.
//
// The client puts no bound of its own on the number of attempts, so a policy
// that always returns true loops until the context is done. Use retries to stop
// after a budget, and honour ctx while waiting.
//
// apiErr is never nil. When the API answered with something the client could not
// parse as an APIError, it is a synthetic one carrying the status code and the
// raw message. When the request never reached the API at all — a transport
// error, a cancelled context — StatusCode is 0.
type RetryPolicy func(ctx context.Context, req *http.Request, apiErr *APIError, retries int) bool

// NotRetryablePolicy never retries. This is the default.
func NotRetryablePolicy(_ context.Context, _ *http.Request, _ *APIError, _ int) bool {
	return false
}

// ServerErrorRetryablePolicy retries server errors (HTTP 5xx) after
// ServerErrorRetryDelay. Client errors are not retried: sending the same
// malformed or unauthorized request again would fail the same way.
//
// It has no attempt budget of its own, and relies on the context to end: give
// the call a deadline, or wrap it in a policy of your own that stops on retries.
func ServerErrorRetryablePolicy(ctx context.Context, _ *http.Request, apiErr *APIError, _ int) bool {
	if apiErr == nil || apiErr.StatusCode < http.StatusInternalServerError || apiErr.StatusCode > 599 {
		return false
	}

	timer := time.NewTimer(ServerErrorRetryDelay)
	defer timer.Stop()

	if ctx == nil {
		<-timer.C

		return true
	}

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
