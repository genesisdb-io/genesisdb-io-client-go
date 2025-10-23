package genesisdb

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	APIURL     string
	APIVersion string
	AuthToken  string
}

type Genesisdb struct {
	config   *Config
	client   *http.Client
}

type RFC3339Time time.Time

func (t *RFC3339Time) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "" {
		*t = RFC3339Time(time.Time{})
		return nil
	}
	parsedTime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = RFC3339Time(parsedTime)
	return nil
}

func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))), nil
}

func (t RFC3339Time) Time() time.Time {
	return time.Time(t)
}

type Event struct {
	ID              string                 `json:"id,omitempty"`
	Source          string                 `json:"source,omitempty"`
	Subject         string                 `json:"subject"`
	Type            string                 `json:"type"`
	Time            RFC3339Time            `json:"time,omitempty"`
	Data            interface{}            `json:"data"`
	DataContentType string                 `json:"datacontenttype,omitempty"`
	SpecVersion     string                 `json:"specversion,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

type Precondition struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type CommitRequest struct {
	Events        []Event         `json:"events"`
	Preconditions []Precondition  `json:"preconditions,omitempty"`
}

type StreamOptions struct {
	LowerBound            string `json:"lowerBound,omitempty"`
	IncludeLowerBoundEvent bool   `json:"includeLowerBoundEvent,omitempty"`
	LatestByEventType     string `json:"latestByEventType,omitempty"`
}

type StreamRequest struct {
	Subject string         `json:"subject"`
	Options *StreamOptions `json:"options,omitempty"`
}

func NewClient(config *Config) (*Genesisdb, error) {
	if config.APIURL == "" {
		return nil, fmt.Errorf("APIURL is required")
	}
	if config.APIVersion == "" {
		return nil, fmt.Errorf("APIVersion is required")
	}
	if config.AuthToken == "" {
		return nil, fmt.Errorf("AuthToken is required")
	}

	return &Genesisdb{
		config: config,
		client: &http.Client{},
	}, nil
}

func (es *Genesisdb) StreamEvents(subject string, options *StreamOptions) ([]Event, error) {
	url := fmt.Sprintf("%s/api/%s/stream", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

	requestBody := StreamRequest{
		Subject: subject,
		Options: options,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")
	req.Header.Set("User-Agent", "genesisdb-sdk-go")

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var events []Event
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("error parsing event JSON: %w", err)
		}

		if event.ID == "" {
			event.ID = uuid.New().String()
		}
		if event.Source == "" {
			event.Source = es.config.APIURL
		}
		if event.DataContentType == "" {
			event.DataContentType = "application/json"
		}
		if event.SpecVersion == "" {
			event.SpecVersion = "1.0"
		}
		if event.Time == RFC3339Time(time.Time{}) {
			now := time.Now().UTC()
			event.Time = RFC3339Time(now)
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return events, nil
}

func (es *Genesisdb) CommitEvents(events []Event) error {
	return es.CommitEventsWithPreconditions(events, nil)
}

func (es *Genesisdb) CommitEventsWithPreconditions(events []Event, preconditions []Precondition) error {
	return es.CommitEventsWithOptions(events, preconditions)
}

func (es *Genesisdb) CommitEventsWithOptions(events []Event, preconditions []Precondition) error {
	url := fmt.Sprintf("%s/api/%s/commit", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

	for i := range events {
		if events[i].ID == "" {
			events[i].ID = uuid.New().String()
		}
		if events[i].Source == "" {
			events[i].Source = es.config.APIURL
		}
		if events[i].DataContentType == "" {
			events[i].DataContentType = "application/json"
		}
		if events[i].SpecVersion == "" {
			events[i].SpecVersion = "1.0"
		}
		if events[i].Time == RFC3339Time(time.Time{}) {
			now := time.Now().UTC()
			events[i].Time = RFC3339Time(now)
		}
	}

	commitRequest := CommitRequest{
		Events: events,
	}
	if preconditions != nil {
		commitRequest.Preconditions = preconditions
	}

	requestBody, err := json.Marshal(commitRequest)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "genesisdb-sdk-go")

	resp, err := es.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (es *Genesisdb) EraseData(subject string) error {
	url := fmt.Sprintf("%s/api/%s/erase", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

	requestBody, err := json.Marshal(map[string]string{"subject": subject})
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "genesisdb-sdk-go")

	resp, err := es.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (es *Genesisdb) Q(query string) ([]interface{}, error) {
	url := fmt.Sprintf("%s/api/%s/q", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

	requestBody, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")
	req.Header.Set("User-Agent", "genesisdb-sdk-go")

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var results []interface{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var result interface{}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return nil, fmt.Errorf("error parsing result JSON: %w", err)
		}
		results = append(results, result)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return results, nil
}

// QueryEvents executes a query using the same functionality as the Q method
// query: The query string to execute
// Returns: Array of query results and any error
// Example:
//   results, err := client.QueryEvents(`FROM e IN events WHERE e.type == "io.genesisdb.app.customer-added" ORDER BY e.time DESC TOP 20 PROJECT INTO { subject: e.subject, firstName: e.data.firstName }`)
func (es *Genesisdb) QueryEvents(query string) ([]interface{}, error) {
	return es.Q(query)
}

func (es *Genesisdb) Ping() (string, error) {
	url := fmt.Sprintf("%s/api/%s/status/ping", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
	req.Header.Set("User-Agent", "genesisdb-sdk-go")

	resp, err := es.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(bodyBytes), nil
}

func (es *Genesisdb) Audit() (string, error) {
	url := fmt.Sprintf("%s/api/%s/status/audit", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
	req.Header.Set("User-Agent", "genesisdb-sdk-go")

	resp, err := es.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(bodyBytes), nil
}

func (es *Genesisdb) ObserveEvents(subject string, options *StreamOptions) (<-chan Event, <-chan error) {
	eventChan := make(chan Event, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errorChan)

		url := fmt.Sprintf("%s/api/%s/observe", strings.TrimRight(es.config.APIURL, "/"), es.config.APIVersion)

		requestBody := StreamRequest{
			Subject: subject,
			Options: options,
		}

		requestBodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			errorChan <- fmt.Errorf("error marshaling request: %w", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
		if err != nil {
			errorChan <- fmt.Errorf("error creating request: %w", err)
			return
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.config.AuthToken))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/x-ndjson")
		req.Header.Set("User-Agent", "genesisdb-sdk-go")

		resp, err := es.client.Do(req)
		if err != nil {
			errorChan <- fmt.Errorf("error making request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errorChan <- fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			jsonStr := line
			if strings.HasPrefix(line, "data: ") {
				jsonStr = line[6:]
			}

			// Check if this is an empty payload object with only one key
			var jsonMap map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &jsonMap); err == nil {
				if payload, ok := jsonMap["payload"].(string); ok && payload == "" && len(jsonMap) == 1 {
					continue
				}
			}

			var event Event
			if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
				errorChan <- fmt.Errorf("error parsing event JSON: %w", err)
				continue
			}

			if event.ID == "" {
				event.ID = uuid.New().String()
			}
			if event.Source == "" {
				event.Source = es.config.APIURL
			}
			if event.DataContentType == "" {
				event.DataContentType = "application/json"
			}
			if event.SpecVersion == "" {
				event.SpecVersion = "1.0"
			}
			if event.Time == RFC3339Time(time.Time{}) {
				now := time.Now().UTC()
				event.Time = RFC3339Time(now)
			}

			select {
			case eventChan <- event:
			case <-time.After(5 * time.Second):
				errorChan <- fmt.Errorf("timeout sending event to channel")
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading response: %w", err)
		}
	}()

	return eventChan, errorChan
}
