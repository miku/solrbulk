package solrbulk

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
			err := BulkIndex(tc.docs, tc.options)
			if (err != nil) != tc.expectError {
				t.Errorf("BulkIndex() error = %v, expectError %v", err, tc.expectError)
			}
		})
	}
}
