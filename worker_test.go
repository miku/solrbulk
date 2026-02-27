package solrbulk

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/sethgrid/pester"
)

// TestBulkIndex tests the BulkIndex function
func TestBulkIndex(t *testing.T) {
	tests := []struct {
		name            string
		docs            []string
		options         Options
		serverStatus    int
		responseBody    string
		expectError     bool
		validateRequest func(t *testing.T, r *http.Request)
	}{
		{
			name: "index success",
			docs: []string{
				`{"id":"1","title":"Test Document 1"}`,
				`{"id":"2","title":"Test Document 2"}`,
			},
			options: Options{
				BatchSize:                100,
				CommitSize:               100,
				Verbose:                  false,
				UpdateRequestHandlerName: "/update/json",
			},
			serverStatus: http.StatusOK,
			responseBody: `{"responseHeader":{"status":0,"QTime":5}}`,
			expectError:  false,
			validateRequest: func(t *testing.T, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}

				bodyStr := string(body)
				if !strings.Contains(bodyStr, `"id":"1"`) || !strings.Contains(bodyStr, `"id":"2"`) {
					t.Errorf("Request body missing expected documents: %s", bodyStr)
				}

				if !strings.HasPrefix(bodyStr, "[") || !strings.HasSuffix(bodyStr, "]\n") {
					t.Errorf("Request body not properly formatted: %s", bodyStr)
				}
			},
		},
		{
			name: "skip empty docs",
			docs: []string{
				`{"id":"1","title":"Test Document 1"}`,
				"",
				"  ",
				`{"id":"2","title":"Test Document 2"}`,
			},
			options: Options{
				BatchSize:                100,
				CommitSize:               100,
				Verbose:                  false,
				UpdateRequestHandlerName: "/update/json",
			},
			serverStatus: http.StatusOK,
			responseBody: `{"responseHeader":{"status":0,"QTime":5}}`,
			expectError:  false,
			validateRequest: func(t *testing.T, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}

				bodyStr := string(body)
				count := strings.Count(bodyStr, `"id"`)
				if count != 2 {
					t.Errorf("Expected 2 documents after filtering, got approximately %d", count)
				}
			},
		},
		{
			name: "server error",
			docs: []string{
				`{"id":"1","title":"Test Document 1"}`,
			},
			options: Options{
				BatchSize:                100,
				CommitSize:               100,
				Verbose:                  false,
				UpdateRequestHandlerName: "/update/json",
			},
			serverStatus:    http.StatusInternalServerError,
			responseBody:    `{"responseHeader":{"status":1,"QTime":5,"errors":[{"id":"doc1","type":"error","message":"Error indexing document"}]}}`,
			expectError:     true,
			validateRequest: nil,
		},
		{
			name: "server error with tolerant handler",
			docs: []string{
				`{"id":"1","title":"Test Document 1"}`,
			},
			options: Options{
				BatchSize:                100,
				CommitSize:               100,
				Verbose:                  false,
				UpdateRequestHandlerName: "/update/json",
			},
			serverStatus:    http.StatusOK,
			responseBody:    `{"responseHeader":{"status":1,"QTime":5,"errors":[{"id":"doc1","type":"error","message":"Error indexing document"}]}}`,
			expectError:     false,
			validateRequest: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != tc.options.UpdateRequestHandlerName {
					t.Errorf("Expected path %s, got %s", tc.options.UpdateRequestHandlerName, r.URL.Path)
				}
				if tc.validateRequest != nil {
					tc.validateRequest(t, r)
				}
				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			tc.options.Server = server.URL
			client := pester.New()
			client.MaxRetries = 0
			err := BulkIndex(tc.docs, tc.options, client)
			if (err != nil) != tc.expectError {
				t.Errorf("BulkIndex() error = %v, expectError %v", err, tc.expectError)
			}
		})
	}
}

func TestBulkIndexRetryTransientFailure(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			// Close connection without response to simulate connection refused.
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("server does not support hijacking")
			}
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1}}`))
	}))
	defer server.Close()
	options := Options{
		BatchSize:                100,
		CommitSize:               100,
		UpdateRequestHandlerName: "/update",
		Server:                   server.URL,
	}
	client := pester.New()
	client.MaxRetries = 5
	client.Backoff = pester.LinearBackoff
	err := BulkIndex([]string{`{"id":"1"}`}, options, client)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	got := atomic.LoadInt32(&attempts)
	if got < 3 {
		t.Errorf("expected at least 3 attempts (2 failures + 1 success), got %d", got)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient(Options{MaxRetries: 7, RetryWaitSeconds: 3})
	if c.MaxRetries != 7 {
		t.Errorf("expected MaxRetries=7, got %d", c.MaxRetries)
	}
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient(Options{})
	if c.MaxRetries != 10 {
		t.Errorf("expected default MaxRetries=10, got %d", c.MaxRetries)
	}
}
