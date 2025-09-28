package genesisdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Custom RoundTripper for mocking
type mockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("")))}, nil
}

func TestNewClient_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid configuration",
			config: &Config{
				APIURL:     "http://localhost:8080",
				APIVersion: "v1",
				AuthToken:  "test-token",
			},
			wantErr: false,
		},
		{
			name: "Missing APIURL",
			config: &Config{
				APIVersion: "v1",
				AuthToken:  "test-token",
			},
			wantErr: true,
			errMsg:  "APIURL is required",
		},
		{
			name: "Missing APIVersion",
			config: &Config{
				APIURL:    "http://localhost:8080",
				AuthToken: "test-token",
			},
			wantErr: true,
			errMsg:  "APIVersion is required",
		},
		{
			name: "Missing AuthToken",
			config: &Config{
				APIURL:     "http://localhost:8080",
				APIVersion: "v1",
			},
			wantErr: true,
			errMsg:  "AuthToken is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("NewClient() error = %v, want %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() should return a client")
			}
		})
	}
}

func TestStreamEvents_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Successful stream", func(t *testing.T) {
		event1 := Event{
			ID:      "1",
			Source:  "test",
			Subject: "/test",
			Type:    "test.event",
			Data:    map[string]interface{}{"message": "test1"},
		}
		event2 := Event{
			ID:      "2",
			Source:  "test",
			Subject: "/test",
			Type:    "test.event",
			Data:    map[string]interface{}{"message": "test2"},
		}

		event1JSON, _ := json.Marshal(event1)
		event2JSON, _ := json.Marshal(event2)
		responseBody := fmt.Sprintf("%s\n%s\n", string(event1JSON), string(event2JSON))

		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/stream" {
					t.Errorf("Unexpected path: %s", req.URL.Path)
				}
				if req.Method != "POST" {
					t.Errorf("Unexpected method: %s", req.Method)
				}
				if auth := req.Header.Get("Authorization"); auth != "Bearer test-token" {
					t.Errorf("Unexpected authorization: %s", auth)
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(responseBody)),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events, err := client.StreamEvents("/test", nil)
		if err != nil {
			t.Fatalf("StreamEvents() error = %v", err)
		}
		if len(events) != 2 {
			t.Errorf("StreamEvents() got %d events, want 2", len(events))
		}
	})

	t.Run("With options", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				var requestBody StreamRequest
				json.Unmarshal(body, &requestBody)

				if requestBody.Options == nil {
					t.Error("Expected options in request body")
				}
				if requestBody.Options.LowerBound != "123" {
					t.Errorf("Expected lowerBound to be 123, got %s", requestBody.Options.LowerBound)
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		options := &StreamOptions{
			LowerBound: "123",
			IncludeLowerBoundEvent: true,
		}
		_, err := client.StreamEvents("/test", options)
		if err != nil {
			t.Fatalf("StreamEvents() error = %v", err)
		}
	})

	t.Run("API error", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(strings.NewReader("Server error")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		_, err := client.StreamEvents("/test", nil)
		if err == nil {
			t.Error("StreamEvents() should return error for API error")
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("Error should contain status code, got: %v", err)
		}
	})

	t.Run("Empty response", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events, err := client.StreamEvents("/test", nil)
		if err != nil {
			t.Fatalf("StreamEvents() error = %v", err)
		}
		if len(events) != 0 {
			t.Errorf("StreamEvents() should return empty slice for empty response")
		}
	})

	t.Run("Auto-populate missing fields", func(t *testing.T) {
		eventWithoutFields := map[string]interface{}{
			"subject": "/test",
			"type":    "test.event",
			"data":    map[string]interface{}{"message": "test"},
		}
		eventJSON, _ := json.Marshal(eventWithoutFields)

		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(eventJSON) + "\n")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events, err := client.StreamEvents("/test", nil)
		if err != nil {
			t.Fatalf("StreamEvents() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(events))
		}

		if events[0].ID == "" {
			t.Error("Event ID should be auto-populated")
		}
		if events[0].Source == "" {
			t.Error("Event Source should be auto-populated")
		}
		if events[0].DataContentType == "" {
			t.Error("Event DataContentType should be auto-populated")
		}
		if events[0].SpecVersion == "" {
			t.Error("Event SpecVersion should be auto-populated")
		}
	})
}

func TestCommitEvents_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Successful commit", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/commit" {
					t.Errorf("Unexpected path: %s", req.URL.Path)
				}
				if req.Method != "POST" {
					t.Errorf("Unexpected method: %s", req.Method)
				}

				body, _ := io.ReadAll(req.Body)
				var commitReq CommitRequest
				json.Unmarshal(body, &commitReq)

				if len(commitReq.Events) != 1 {
					t.Errorf("Expected 1 event, got %d", len(commitReq.Events))
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events := []Event{
			{
				Source:  "test",
				Subject: "/test",
				Type:    "test.event",
				Data:    map[string]interface{}{"message": "test"},
			},
		}

		err := client.CommitEvents(events)
		if err != nil {
			t.Fatalf("CommitEvents() error = %v", err)
		}
	})

	t.Run("With preconditions", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				var commitReq CommitRequest
				json.Unmarshal(body, &commitReq)

				if len(commitReq.Preconditions) != 1 {
					t.Errorf("Expected 1 precondition, got %d", len(commitReq.Preconditions))
				}
				if commitReq.Preconditions[0].Type != "isSubjectNew" {
					t.Errorf("Expected precondition type 'isSubjectNew', got %s", commitReq.Preconditions[0].Type)
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events := []Event{
			{
				Source:  "test",
				Subject: "/test",
				Type:    "test.event",
				Data:    map[string]interface{}{"message": "test"},
			},
		}

		preconditions := []Precondition{
			{
				Type: "isSubjectNew",
				Payload: map[string]interface{}{
					"subject": "/test",
				},
			},
		}

		err := client.CommitEventsWithPreconditions(events, preconditions)
		if err != nil {
			t.Fatalf("CommitEventsWithPreconditions() error = %v", err)
		}
	})

	t.Run("With options", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				var commitReq CommitRequest
				json.Unmarshal(body, &commitReq)

				if commitReq.Events[0].Options == nil {
					t.Error("Expected options in event")
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events := []Event{
			{
				Source:  "test",
				Subject: "/test",
				Type:    "test.event",
				Data:    map[string]interface{}{"message": "test"},
				Options: map[string]interface{}{"storeDataAsReference": true},
			},
		}

		err := client.CommitEvents(events)
		if err != nil {
			t.Fatalf("CommitEvents() error = %v", err)
		}
	})

	t.Run("API error", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 400,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader("Invalid request")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		events := []Event{
			{
				Source:  "test",
				Subject: "/test",
				Type:    "test.event",
				Data:    map[string]interface{}{"message": "test"},
			},
		}

		err := client.CommitEvents(events)
		if err == nil {
			t.Error("CommitEvents() should return error for API error")
		}
		if !strings.Contains(err.Error(), "400") {
			t.Errorf("Error should contain status code, got: %v", err)
		}
	})
}

func TestEraseData_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Successful erase", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/erase" {
					t.Errorf("Unexpected path: %s", req.URL.Path)
				}

				body, _ := io.ReadAll(req.Body)
				var eraseReq map[string]string
				json.Unmarshal(body, &eraseReq)

				if eraseReq["subject"] != "/test/subject" {
					t.Errorf("Expected subject '/test/subject', got %s", eraseReq["subject"])
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		err := client.EraseData("/test/subject")
		if err != nil {
			t.Fatalf("EraseData() error = %v", err)
		}
	})

	t.Run("API error", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Body:       io.NopCloser(strings.NewReader("Subject not found")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		err := client.EraseData("/test/subject")
		if err == nil {
			t.Error("EraseData() should return error for API error")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("Error should contain status code, got: %v", err)
		}
	})
}

func TestQ_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Successful query", func(t *testing.T) {
		result1 := map[string]interface{}{"id": "1", "name": "Result 1"}
		result2 := map[string]interface{}{"id": "2", "name": "Result 2"}

		result1JSON, _ := json.Marshal(result1)
		result2JSON, _ := json.Marshal(result2)
		responseBody := fmt.Sprintf("%s\n%s\n", string(result1JSON), string(result2JSON))

		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/q" {
					t.Errorf("Unexpected path: %s", req.URL.Path)
				}

				body, _ := io.ReadAll(req.Body)
				var queryReq map[string]string
				json.Unmarshal(body, &queryReq)

				if queryReq["query"] != "FROM e IN events" {
					t.Errorf("Unexpected query: %s", queryReq["query"])
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(responseBody)),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		results, err := client.Q("FROM e IN events")
		if err != nil {
			t.Fatalf("Q() error = %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Q() got %d results, want 2", len(results))
		}
	})

	t.Run("Empty results", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		results, err := client.Q("FROM e IN events WHERE false")
		if err != nil {
			t.Fatalf("Q() error = %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Q() should return empty slice for empty response")
		}
	})

	t.Run("API error", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 400,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(strings.NewReader("Invalid query")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		_, err := client.Q("INVALID QUERY")
		if err == nil {
			t.Error("Q() should return error for API error")
		}
		if !strings.Contains(err.Error(), "400") {
			t.Errorf("Error should contain status code, got: %v", err)
		}
	})
}

func TestQueryEvents_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	mockTransport := &mockRoundTripper{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			result := map[string]interface{}{"id": "1", "name": "Result"}
			resultJSON, _ := json.Marshal(result)

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(resultJSON) + "\n")),
			}, nil
		},
	}

	client, _ := NewClient(config)
	client.client.Transport = mockTransport

	results, err := client.QueryEvents("FROM e IN events")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("QueryEvents() got %d results, want 1", len(results))
	}
}

func TestPing_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Successful ping", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/status/ping" {
					t.Errorf("Unexpected path: %s", req.URL.Path)
				}
				if req.Method != "GET" {
					t.Errorf("Unexpected method: %s", req.Method)
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("pong")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		response, err := client.Ping()
		if err != nil {
			t.Fatalf("Ping() error = %v", err)
		}
		if response != "pong" {
			t.Errorf("Ping() = %s, want 'pong'", response)
		}
	})

	t.Run("API error", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 503,
					Status:     "503 Service Unavailable",
					Body:       io.NopCloser(strings.NewReader("Service unavailable")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		_, err := client.Ping()
		if err == nil {
			t.Error("Ping() should return error for API error")
		}
		if !strings.Contains(err.Error(), "503") {
			t.Errorf("Error should contain status code, got: %v", err)
		}
	})
}

func TestAudit_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Successful audit", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/status/audit" {
					t.Errorf("Unexpected path: %s", req.URL.Path)
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("Audit successful")),
				}, nil
			},
		}

		client, _ := NewClient(config)
		client.client.Transport = mockTransport

		response, err := client.Audit()
		if err != nil {
			t.Fatalf("Audit() error = %v", err)
		}
		if response != "Audit successful" {
			t.Errorf("Audit() = %s, want 'Audit successful'", response)
		}
	})
}

func TestObserveEvents_Mock(t *testing.T) {
	config := &Config{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "test-token",
	}

	t.Run("Stream events", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/observe" {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}

			event := Event{
				ID:      "1",
				Source:  "test",
				Subject: "/test",
				Type:    "test.event",
				Data:    map[string]interface{}{"message": "test"},
			}
			eventJSON, _ := json.Marshal(event)

			w.WriteHeader(200)
			w.Write([]byte(string(eventJSON) + "\n"))
		}))
		defer server.Close()

		config.APIURL = server.URL
		client, _ := NewClient(config)

		eventChan, errorChan := client.ObserveEvents("/test", nil)

		select {
		case event := <-eventChan:
			if event.Type != "test.event" {
				t.Errorf("Unexpected event type: %s", event.Type)
			}
		case err := <-errorChan:
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for event")
		}
	})

	t.Run("Handle SSE format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			event := Event{
				ID:      "1",
				Source:  "test",
				Subject: "/test",
				Type:    "test.sse",
				Data:    map[string]interface{}{"message": "sse"},
			}
			eventJSON, _ := json.Marshal(event)

			w.WriteHeader(200)
			w.Write([]byte("data: " + string(eventJSON) + "\n"))
		}))
		defer server.Close()

		config.APIURL = server.URL
		client, _ := NewClient(config)

		eventChan, errorChan := client.ObserveEvents("/test", nil)

		select {
		case event := <-eventChan:
			if event.Type != "test.sse" {
				t.Errorf("Unexpected event type: %s", event.Type)
			}
		case err := <-errorChan:
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for event")
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		config.APIURL = server.URL
		client, _ := NewClient(config)

		eventChan, errorChan := client.ObserveEvents("/test", nil)

		select {
		case <-eventChan:
			t.Fatal("Should not receive events on error")
		case err := <-errorChan:
			if err == nil {
				t.Fatal("Expected error")
			}
			if !strings.Contains(err.Error(), "500") {
				t.Errorf("Error should contain status code, got: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for error")
		}
	})
}

func TestRFC3339Time_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "Valid time",
			time:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: `"2024-01-01T12:00:00Z"`,
		},
		{
			name:     "Zero time",
			time:     time.Time{},
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfcTime := RFC3339Time(tt.time)
			data, err := rfcTime.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", string(data), tt.expected)
			}
		})
	}
}

func TestRFC3339Time_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "Valid time",
			json:    `"2024-01-01T12:00:00Z"`,
			wantErr: false,
		},
		{
			name:    "Null time",
			json:    `"null"`,
			wantErr: false,
		},
		{
			name:    "Empty string",
			json:    `""`,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			json:    `"not-a-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rfcTime RFC3339Time
			err := rfcTime.UnmarshalJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRFC3339Time_Time(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	rfcTime := RFC3339Time(now)

	if !rfcTime.Time().Equal(now) {
		t.Errorf("Time() = %v, want %v", rfcTime.Time(), now)
	}
}


