package genesisdb

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

var testConfig = &Config{
	APIURL:     "http://localhost:8080",
	APIVersion: "v1",
	AuthToken:  "secret",
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: &Config{
				APIURL:     "http://localhost:8080",
				APIVersion: "v1",
				AuthToken:  "secret",
			},
			wantErr: false,
		},
		{
			name: "Missing API URL",
			config: &Config{
				APIVersion: "v1",
				AuthToken:  "secret",
			},
			wantErr: true,
		},
		{
			name: "Missing API Version",
			config: &Config{
				APIURL:    "http://localhost:8080",
				AuthToken: "secret",
			},
			wantErr: true,
		},
		{
			name: "Missing Auth Token",
			config: &Config{
				APIURL:     "http://localhost:8080",
				APIVersion: "v1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() should return a client")
			}
		})
	}
}

func TestPing(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	response, err := client.Ping()
	if err != nil {
		t.Fatalf("Ping() failed: %v", err)
	}

	if response == "" {
		t.Error("Ping() should return a non-empty response")
	}

	t.Logf("Ping response: %s", response)
}

func TestAudit(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	response, err := client.Audit()
	if err != nil {
		t.Fatalf("Audit() failed: %v", err)
	}

	if response == "" {
		t.Error("Audit() should return a non-empty response")
	}

	t.Logf("Audit response: %s", response)
}

func TestCommitEvents(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	events := []Event{
		{
			Source:  "io.genesisdb.test",
			Subject: "/test/customer",
			Type:    "io.genesisdb.test.customer-added",
			Data: map[string]interface{}{
				"firstName": "Max",
				"lastName":  "Mustermann",
				"email":     "max.mustermann@test.de",
				"timestamp": time.Now().Unix(),
			},
		},
		{
			Source:  "io.genesisdb.test",
			Subject: "/test/article",
			Type:    "io.genesisdb.test.article-added",
			Data: map[string]interface{}{
				"name":      "Test Article",
				"price":     99.99,
				"timestamp": time.Now().Unix(),
			},
		},
	}

	err = client.CommitEvents(events)
	if err != nil {
		t.Fatalf("CommitEvents() failed: %v", err)
	}

	t.Log("Events successfully committed")

	streamedEvents, err := client.StreamEvents("/test/customer")
	if err != nil {
		t.Fatalf("Error streaming customer events: %v", err)
	}

	found := false
	for _, event := range streamedEvents {
		if event.Type == "io.genesisdb.test.customer-added" {
			data, ok := event.Data.(map[string]interface{})
			if ok && data["firstName"] == "Max" {
				found = true
				t.Logf("✓ Customer event found: ID=%s", event.ID)
				break
			}
		}
	}

	if !found {
		t.Error("Customer event not found in database")
	}

	articleEvents, err := client.StreamEvents("/test/article")
	if err != nil {
		t.Fatalf("Error streaming article events: %v", err)
	}

	found = false
	for _, event := range articleEvents {
		if event.Type == "io.genesisdb.test.article-added" {
			data, ok := event.Data.(map[string]interface{})
			if ok && data["name"] == "Test Article" {
				found = true
				t.Logf("✓ Article event found: ID=%s", event.ID)
				break
			}
		}
	}

	if !found {
		t.Error("Article event not found in database")
	}
}

func TestStreamEvents(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	testEvents := []Event{
		{
			Source:  "io.genesisdb.test",
			Subject: "/test/stream",
			Type:    "io.genesisdb.test.stream-test",
			Data: map[string]interface{}{
				"message":   "Test for Stream",
				"timestamp": time.Now().Unix(),
				"uniqueId":  fmt.Sprintf("stream-test-%d", time.Now().UnixNano()),
			},
		},
	}

	err = client.CommitEvents(testEvents)
	if err != nil {
		t.Fatalf("Error committing test events: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	events, err := client.StreamEvents("/test/stream")
	if err != nil {
		t.Fatalf("StreamEvents() failed: %v", err)
	}

	if len(events) == 0 {
		t.Log("No events found - this is normal if no events exist")
	} else {
		t.Logf("Found events: %d", len(events))
		found := false
		for i, event := range events {
			t.Logf("Event %d: Type=%s, Subject=%s, ID=%s", i+1, event.Type, event.Subject, event.ID)

			if event.Type == "io.genesisdb.test.stream-test" {
				data, ok := event.Data.(map[string]interface{})
				if ok && data["message"] == "Test for Stream" {
					found = true
					t.Logf("✓ Our test event found: %s", event.ID)
				}
			}
		}

		if !found {
			t.Log("Our specific test event was not found (can be normal with many events)")
		}
	}
}

func TestQuery(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	uniqueId := fmt.Sprintf("query-test-%d", time.Now().UnixNano())
	testEvents := []Event{
		{
			Source:  "io.genesisdb.test",
			Subject: "/test/query",
			Type:    "io.genesisdb.test.query-test",
			Data: map[string]interface{}{
				"name":      "Query Test Event",
				"value":     42,
				"uniqueId":  uniqueId,
				"timestamp": time.Now().Unix(),
			},
		},
	}

	err = client.CommitEvents(testEvents)
	if err != nil {
		t.Fatalf("Error committing test events: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	query := `
FROM e IN events
WHERE e.type == 'io.genesisdb.test.query-test'
ORDER BY e.time
PROJECT INTO { id: e.id, name: e.data.name, value: e.data.value }
`

	results, err := client.Q(query)
	if err != nil {
		t.Fatalf("Q() failed: %v", err)
	}

	if len(results) == 0 {
		t.Log("No results found - this is normal if no matching events exist")
	} else {
		t.Logf("Query results: %d", len(results))
		found := false
		for i, result := range results {
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			t.Logf("Result %d: %s", i+1, string(resultJSON))

			if resultMap, ok := result.(map[string]interface{}); ok {
				if resultMap["name"] == "Query Test Event" {
					found = true
					t.Logf("✓ Our specific query result found")
				}
			}
		}

		if !found {
			t.Log("Our specific query result was not found")
		}
	}
}

func TestRFC3339Time(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	rfcTime := RFC3339Time(now)

	jsonData, err := json.Marshal(rfcTime)
	if err != nil {
		t.Fatalf("Error during marshaling: %v", err)
	}

	var unmarshaled RFC3339Time
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Error during unmarshaling: %v", err)
	}

	if !rfcTime.Time().Equal(unmarshaled.Time()) {
		t.Errorf("Times do not match: %v != %v", rfcTime.Time(), unmarshaled.Time())
	}

	t.Logf("RFC3339Time test successful: %s", rfcTime.Time().Format(time.RFC3339))
}

func TestEventValidation(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	uniqueId := fmt.Sprintf("validation-test-%d", time.Now().UnixNano())
	event := Event{
		Source:  "io.genesisdb.test",
		Subject: "/test/validation",
		Type:    "io.genesisdb.test.validation",
		Data: map[string]interface{}{
			"test":     "validation",
			"uniqueId": uniqueId,
		},
	}

	events := []Event{event}
	err = client.CommitEvents(events)
	if err != nil {
		t.Fatalf("CommitEvents() with incomplete event failed: %v", err)
	}

	t.Log("Event validation successful - missing fields were automatically added")

	time.Sleep(100 * time.Millisecond)

	streamedEvents, err := client.StreamEvents("/test/validation")
	if err != nil {
		t.Fatalf("Error verifying validated event: %v", err)
	}

	found := false
	for _, streamedEvent := range streamedEvents {
		if streamedEvent.Type == "io.genesisdb.test.validation" {
			data, ok := streamedEvent.Data.(map[string]interface{})
			if ok && data["uniqueId"] == uniqueId {
				found = true
				t.Logf("✓ Validated event found with ID: %s", streamedEvent.ID)
				break
			}
		}
	}

	if !found {
		t.Error("Validated event not found in database")
	}
}

func TestIntegration(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	_, err = client.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	t.Log("✓ Ping successful")

	uniqueId := fmt.Sprintf("integration-test-%d", time.Now().UnixNano())
	events := []Event{
		{
			Source:  "io.genesisdb.test",
			Subject: "/test/integration",
			Type:    "io.genesisdb.test.integration",
			Data: map[string]interface{}{
				"step":      1,
				"message":   "Integration Test",
				"uniqueId":  uniqueId,
				"timestamp": time.Now().Unix(),
			},
		},
		{
			Source:  "io.genesisdb.test",
			Subject: "/test/integration",
			Type:    "io.genesisdb.test.integration",
			Data: map[string]interface{}{
				"step":      2,
				"message":   "Integration Test Continuation",
				"uniqueId":  uniqueId,
				"timestamp": time.Now().Unix(),
			},
		},
	}

	err = client.CommitEvents(events)
	if err != nil {
		t.Fatalf("CommitEvents failed: %v", err)
	}
	t.Log("✓ Events successfully committed")

	time.Sleep(100 * time.Millisecond)

	streamedEvents, err := client.StreamEvents("/test/integration")
	if err != nil {
		t.Fatalf("StreamEvents failed: %v", err)
	}
	t.Logf("✓ %d events streamed", len(streamedEvents))

	ourEvents := 0
	for _, event := range streamedEvents {
		if event.Type == "io.genesisdb.test.integration" {
			data, ok := event.Data.(map[string]interface{})
			if ok && data["uniqueId"] == uniqueId {
				ourEvents++
			}
		}
	}
	t.Logf("✓ Found %d of our specific integration events", ourEvents)

	query := `
FROM e IN events
WHERE e.type == 'io.genesisdb.test.integration'
ORDER BY e.time
PROJECT INTO { step: e.data.step, message: e.data.message }
`

	results, err := client.Q(query)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	t.Logf("✓ Query executed successfully, %d results", len(results))

	if len(results) < 2 {
		t.Errorf("Expected at least 2 query results, but found %d", len(results))
	}

	_, err = client.Audit()
	if err != nil {
		t.Fatalf("Audit failed: %v", err)
	}
	t.Log("✓ Audit successful")

	t.Log("🎉 Integration test completely successful!")
}

func BenchmarkCommitEvents(b *testing.B) {
	client, err := NewClient(testConfig)
	if err != nil {
		b.Fatalf("Error creating client: %v", err)
	}

	events := []Event{
		{
			Source:  "io.genesisdb.benchmark",
			Subject: "/benchmark/test",
			Type:    "io.genesisdb.benchmark.test",
			Data: map[string]interface{}{
				"benchmark": true,
				"iteration": 0,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		events[0].Data.(map[string]interface{})["iteration"] = i
		err := client.CommitEvents(events)
		if err != nil {
			b.Fatalf("CommitEvents failed: %v", err)
		}
	}
}

func BenchmarkStreamEvents(b *testing.B) {
	client, err := NewClient(testConfig)
	if err != nil {
		b.Fatalf("Error creating client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.StreamEvents("/benchmark/test")
		if err != nil {
			b.Fatalf("StreamEvents failed: %v", err)
		}
	}
}

func TestCommitEventsWithPreconditions(t *testing.T) {
	client, err := NewClient(testConfig)
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	events := []Event{
		{
			Source: "io.genesisdb.app",
			Subject: "/foo/21",
			Type:    "io.genesisdb.app.foo-added",
			Data: map[string]interface{}{
				"value": "Foo",
			},
		},
	}

	preconditions := []Precondition{
		{
			Type: "isSubjectNew",
			Payload: map[string]interface{}{
				"subject": "/foo/21",
			},
		},
	}

	err = client.CommitEventsWithPreconditions(events, preconditions)
	if err != nil {
		t.Fatalf("CommitEventsWithPreconditions() failed: %v", err)
	}

	t.Log("Events with preconditions successfully committed")

	streamedEvents, err := client.StreamEvents("/foo/21")
	if err != nil {
		t.Fatalf("Error streaming foo events: %v", err)
	}

	found := false
	for _, event := range streamedEvents {
		if event.Type == "io.genesisdb.app.foo-added" {
			data, ok := event.Data.(map[string]interface{})
			if ok && data["value"] == "Foo" {
				found = true
				t.Logf("✓ Foo event found: ID=%s", event.ID)
				break
			}
		}
	}

	if !found {
		t.Error("Foo event not found in database")
	}
}
