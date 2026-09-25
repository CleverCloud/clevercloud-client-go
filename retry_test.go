package client_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.clever-cloud.dev/client"
)

type payload struct {
	Name string `json:"name"`
}

// serverFailing answers with status for the first failures calls, then 200.
func serverFailing(t *testing.T, status int, failures int32, calls *int32, bodies *[]string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(calls, 1)

		if bodies != nil {
			body, _ := io.ReadAll(r.Body)
			*bodies = append(*bodies, string(body))
		}

		if attempt <= failures {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"id":4004,"message":"A server error occurred","type":"error"}`))

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"ok"}`))
	}))
	t.Cleanup(srv.Close)

	return srv
}

func TestRetryPolicy_DefaultDoesNotRetry(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := serverFailing(t, http.StatusInternalServerError, 1, &calls, nil)

	cc := client.New(client.WithEndpoint(srv.URL))
	res := client.Get[payload](context.Background(), cc, "/any")

	if !res.HasError() {
		t.Fatal("expected the request to fail")
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 call with the default policy, got %d", got)
	}
}

func TestRetryPolicy_ServerErrorRetries(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := serverFailing(t, http.StatusInternalServerError, 1, &calls, nil)

	cc := client.New(
		client.WithEndpoint(srv.URL),
		client.WithRetryPolicy(client.ServerErrorRetryablePolicy),
	)
	res := client.Get[payload](context.Background(), cc, "/any")

	if res.HasError() {
		t.Fatalf("expected the retry to succeed, got %s", res.Error())
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}

	if got := res.Payload().Name; got != "ok" {
		t.Fatalf("expected the payload of the successful attempt, got %q", got)
	}
}

func TestRetryPolicy_ServerErrorDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := serverFailing(t, http.StatusBadRequest, 1, &calls, nil)

	cc := client.New(
		client.WithEndpoint(srv.URL),
		client.WithRetryPolicy(client.ServerErrorRetryablePolicy),
	)
	res := client.Get[payload](context.Background(), cc, "/any")

	if !res.HasError() {
		t.Fatal("expected the request to fail")
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected a 400 not to be retried, got %d calls", got)
	}
}

func TestRetryPolicy_PolicyOwnsItsBudget(t *testing.T) {
	t.Parallel()

	var calls int32
	// Never recovers: only the policy can end the loop.
	srv := serverFailing(t, http.StatusServiceUnavailable, 100, &calls, nil)

	// The client puts no bound of its own on the attempts, so the budget lives
	// in the policy, which is what the retries argument is for.
	budgeted := func(ctx context.Context, req *http.Request, apiErr *client.APIError, retries int) bool {
		if retries >= 2 {
			return false
		}

		return client.ServerErrorRetryablePolicy(ctx, req, apiErr, retries)
	}

	cc := client.New(client.WithEndpoint(srv.URL), client.WithRetryPolicy(budgeted))
	res := client.Get[payload](context.Background(), cc, "/any")

	if !res.HasError() {
		t.Fatal("expected the request to fail")
	}

	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 1 call and 2 retries, got %d calls", got)
	}
}

func TestRetryPolicy_CountsRetries(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := serverFailing(t, http.StatusInternalServerError, 100, &calls, nil)

	seen := []int{}
	counting := func(_ context.Context, _ *http.Request, _ *client.APIError, retries int) bool {
		seen = append(seen, retries)

		return retries < 3
	}

	cc := client.New(client.WithEndpoint(srv.URL), client.WithRetryPolicy(counting))
	client.Get[payload](context.Background(), cc, "/any")

	want := []int{0, 1, 2, 3}
	if len(seen) != len(want) {
		t.Fatalf("expected the policy to be called %d times, got %d: %v", len(want), len(seen), seen)
	}

	for i, retries := range want {
		if seen[i] != retries {
			t.Errorf("call %d got retries=%d, expected %d", i+1, seen[i], retries)
		}
	}
}

func TestRetryPolicy_StopsWhenContextIsDone(t *testing.T) {
	t.Parallel()

	var calls int32
	// Never recovers, and the policy has no budget: only the deadline ends it.
	srv := serverFailing(t, http.StatusInternalServerError, 100, &calls, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*client.ServerErrorRetryDelay)
	defer cancel()

	cc := client.New(
		client.WithEndpoint(srv.URL),
		client.WithRetryPolicy(client.ServerErrorRetryablePolicy),
	)
	res := client.Get[payload](ctx, cc, "/any")

	if !res.HasError() {
		t.Fatal("expected the request to fail")
	}

	if got := atomic.LoadInt32(&calls); got < 1 || got > 4 {
		t.Fatalf("expected the deadline to end the loop after a few attempts, got %d calls", got)
	}
}

func TestRetryPolicy_ReceivesRequestAndStatus(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := serverFailing(t, http.StatusBadGateway, 1, &calls, nil)

	var gotMethod, gotPath string
	var gotStatus int
	var gotMessage string

	policy := func(_ context.Context, req *http.Request, apiErr *client.APIError, _ int) bool {
		gotMethod = req.Method
		gotPath = req.URL.Path
		gotStatus = apiErr.StatusCode
		gotMessage = apiErr.Message

		return false
	}

	cc := client.New(client.WithEndpoint(srv.URL), client.WithRetryPolicy(policy))
	client.Get[payload](context.Background(), cc, "/some/path")

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %q", gotMethod)
	}

	if gotPath != "/some/path" {
		t.Errorf("expected path /some/path, got %q", gotPath)
	}

	if gotStatus != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", gotStatus)
	}

	if gotMessage != "A server error occurred" {
		t.Errorf("expected the API message, got %q", gotMessage)
	}
}

func TestRetryPolicy_ResendsTheBody(t *testing.T) {
	t.Parallel()

	var calls int32
	bodies := []string{}
	srv := serverFailing(t, http.StatusInternalServerError, 1, &calls, &bodies)

	cc := client.New(
		client.WithEndpoint(srv.URL),
		client.WithRetryPolicy(client.ServerErrorRetryablePolicy),
	)
	res := client.Post[payload](context.Background(), cc, "/any", payload{Name: "sent"})

	if res.HasError() {
		t.Fatalf("expected the retry to succeed, got %s", res.Error())
	}

	if len(bodies) != 2 {
		t.Fatalf("expected 2 bodies, got %d", len(bodies))
	}

	for attempt, body := range bodies {
		if body != `{"name":"sent"}` {
			t.Errorf("attempt %d sent %q, expected the payload to be resent", attempt+1, body)
		}
	}
}

func TestRetryPolicy_ServerErrorGivesUpOnCancelledContext(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := serverFailing(t, http.StatusInternalServerError, 100, &calls, nil)

	ctx, cancel := context.WithCancel(context.Background())

	cc := client.New(
		client.WithEndpoint(srv.URL),
		client.WithRetryPolicy(func(ctx context.Context, req *http.Request, apiErr *client.APIError, retries int) bool {
			cancel()

			return client.ServerErrorRetryablePolicy(ctx, req, apiErr, retries)
		}),
	)

	start := time.Now()
	res := client.Get[payload](ctx, cc, "/any")

	if !res.HasError() {
		t.Fatal("expected the request to fail")
	}

	if elapsed := time.Since(start); elapsed >= client.ServerErrorRetryDelay {
		t.Errorf("expected the wait to be cut short, took %s", elapsed)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected no retry after cancellation, got %d calls", got)
	}
}
