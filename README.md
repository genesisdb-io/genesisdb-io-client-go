# Genesis DB Go SDK

This is the official Go SDK for Genesis DB, an awesome and production ready event store database system for building event-driven apps.

## Genesis DB Advantages

* Incredibly fast when reading, fast when writing 🚀
* Easy backup creation and recovery
* [CloudEvents](https://cloudevents.io/) compatible
* GDPR-ready
* Easily accessible via the HTTP interface
* Auditable. Guarantee database consistency
* Logging and metrics for Prometheus
* SQL like query language called Genesis DB Query Language (GDBQL)
* ...

This SDK provides a simple interface to interact with the Genesis DB API.

## Requirements

* Go 1.22 or higher

## Installation

```bash
go get github.com/genesisdb-io/genesisdb-io-client-go
```

## Configuration

The SDK requires the following environment variables to be set:

* `GENESISDB_API_URL`: The URL of the Genesis DB API
* `GENESISDB_API_VERSION`: The version of the API to use
* `GENESISDB_AUTH_TOKEN`: Your authentication token

Alternatively, you can pass these values directly when creating the client:

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

config := &genesisdb.Config{
    APIURL:     "https://your-api-url",
    APIVersion: "v1",
    AuthToken:  "your-auth-token",
}

client, err := genesisdb.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

## Usage

### Streaming Events

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

events, err := client.StreamEvents("/customer")
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event Type: %s, Data: %v\n", event.Type, event.Data)
}
```

### Stream Events from lower bound

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LowerBound:            "2d6d4141-6107-4fb2-905f-445730f4f2a9",
    IncludeLowerBoundEvent: true,
}

events, err := client.StreamEvents("/", options)
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event Type: %s, Data: %v\n", event.Type, event.Data)
}
```

### Stream Events with upper bound

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    UpperBound:            "9f3e4141-7208-4fb2-905f-445730f4f3b1",
    IncludeUpperBoundEvent: false,
}

events, err := client.StreamEvents("/", options)
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event Type: %s, Data: %v\n", event.Type, event.Data)
}
```

### Stream Events with both lower and upper bounds

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LowerBound:            "2d6d4141-6107-4fb2-905f-445730f4f2a9",
    IncludeLowerBoundEvent: true,
    UpperBound:            "9f3e4141-7208-4fb2-905f-445730f4f3b1",
    IncludeUpperBoundEvent: false,
}

events, err := client.StreamEvents("/", options)
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event Type: %s, Data: %v\n", event.Type, event.Data)
}
```

### Stream Events with latest by event type

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LatestByEventType: "io.genesisdb.app.customer-updated",
}

events, err := client.StreamEvents("/", options)
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event Type: %s, Data: %v\n", event.Type, event.Data)
}
```

This feature allows you to stream only the latest event of a specific type for each subject. Useful for getting the current state of entities.

### Observing Events in Real-Time

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

// Start observing events for a subject
eventChan, errorChan := client.ObserveEvents("/customer")

// Listen for events in a goroutine
go func() {
    for {
        select {
        case event := <-eventChan:
            fmt.Printf("Real-time event: Type=%s, Subject=%s, Data=%v\n",
                event.Type, event.Subject, event.Data)
        case err := <-errorChan:
            if err != nil {
                fmt.Printf("Observe error: %v\n", err)
            }
            return
        }
    }
}()

// The observe connection will stay open and stream events as they occur
```

### Observe Events from lower bound (Message queue)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LowerBound:            "2d6d4141-6107-4fb2-905f-445730f4f2a9",
    IncludeLowerBoundEvent: true,
}

// Start observing events for a subject with lower bound
eventChan, errorChan := client.ObserveEvents("/customer", options)

// Listen for events in a goroutine
go func() {
    for {
        select {
        case event := <-eventChan:
            fmt.Printf("Real-time event: Type=%s, Subject=%s, Data=%v\n",
                event.Type, event.Subject, event.Data)
        case err := <-errorChan:
            if err != nil {
                fmt.Printf("Observe error: %v\n", err)
            }
            return
        }
    }
}()

// The observe connection will stay open and stream events as they occur
```

### Observe Events with upper bound (Message queue)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    UpperBound:            "9f3e4141-7208-4fb2-905f-445730f4f3b1",
    IncludeUpperBoundEvent: false,
}

// Start observing events for a subject with upper bound
eventChan, errorChan := client.ObserveEvents("/customer", options)

// Listen for events in a goroutine
go func() {
    for {
        select {
        case event := <-eventChan:
            fmt.Printf("Real-time event: Type=%s, Subject=%s, Data=%v\n",
                event.Type, event.Subject, event.Data)
        case err := <-errorChan:
            if err != nil {
                fmt.Printf("Observe error: %v\n", err)
            }
            return
        }
    }
}()
```

### Observe Events with both bounds (Message queue)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LowerBound:            "2d6d4141-6107-4fb2-905f-445730f4f2a9",
    IncludeLowerBoundEvent: true,
    UpperBound:            "9f3e4141-7208-4fb2-905f-445730f4f3b1",
    IncludeUpperBoundEvent: false,
}

// Start observing events for a subject with both bounds
eventChan, errorChan := client.ObserveEvents("/customer", options)

// Listen for events in a goroutine
go func() {
    for {
        select {
        case event := <-eventChan:
            fmt.Printf("Real-time event: Type=%s, Subject=%s, Data=%v\n",
                event.Type, event.Subject, event.Data)
        case err := <-errorChan:
            if err != nil {
                fmt.Printf("Observe error: %v\n", err)
            }
            return
        }
    }
}()
```

### Observe Latest Events by Event Type (Message queue)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LatestByEventType: "io.genesisdb.app.customer-updated",
}

// Start observing latest events by type
eventChan, errorChan := client.ObserveEvents("/customer", options)

// Listen for events in a goroutine
go func() {
    for {
        select {
        case event := <-eventChan:
            fmt.Printf("Latest event: Type=%s, Subject=%s, Data=%v\n",
                event.Type, event.Subject, event.Data)
        case err := <-errorChan:
            if err != nil {
                fmt.Printf("Observe error: %v\n", err)
            }
            return
        }
    }
}()
```

### Committing Events

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

// Example for creating new events
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.app",
        Subject: "/customer",
        Type:    "io.genesisdb.app.customer-added",
        Data: map[string]interface{}{
            "firstName": "Bruce",
            "lastName": "Wayne",
            "emailAddress": "bruce.wayne@enterprise.wayne"
        },
    },
    {
        Source: "io.genesisdb.app",
        Subject: "/customer",
        Type:    "io.genesisdb.app.customer-added",
        Data: map[string]interface{}{
            "firstName": "Alfred",
            "lastName": "Pennyworth",
            "emailAddress": "alfred.pennyworth@enterprise.wayne"
        },
    },
    {
        Source: "io.genesisdb.store",
        Subject: "/article",
        Type:    "io.genesisdb.store.article-added",
        Data: map[string]interface{}{
            "name": "Tumbler",
            "color": "black",
            "price": 2990000.00
        },
    },
    {
        Source: "io.genesisdb.app",
        Subject: "/customer/fed2902d-0135-460d-8605-263a06308448",
        Type:    "io.genesisdb.app.customer-personaldata-changed",
        Data: map[string]interface{}{
            "firstName": "Angus",
            "lastName": "MacGyver",
            "emailAddress": "angus.macgyer@phoenix.foundation"
        },
    },
}

err := client.CommitEvents(events)
if err != nil {
    log.Fatal(err)
}
```

### Usage of referenced data (GDPR)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

events := []genesisdb.Event{
    {
        Source: "io.genesisdb.app",
        Subject: "/user/456",
        Type:    "io.genesisdb.app.user-created",
        Data: map[string]interface{}{
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
        },
    },
}

options := &genesisdb.CommitOptions{
    StoreDataAsReference: true,
}

err := client.CommitEventsWithOptions(events, options)
if err != nil {
    log.Fatal(err)
}
```

### Deleting referenced data (GDPR)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

err := client.EraseData("/user/456")
if err != nil {
    log.Fatal(err)
}
```

### Committing Events with Preconditions

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

// Example for creating events with preconditions
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.app",
        Subject: "/foo/21",
        Type:    "io.genesisdb.app.foo-added",
        Data: map[string]interface{}{
            "value": "Foo",
        },
    },
}

```

## Preconditions

Preconditions allow you to enforce certain checks on the server before committing events. Genesis DB supports multiple precondition types:

### isSubjectNew
Ensures that a subject is new (has no existing events):

```go
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.app",
        Subject: "/user/456",
        Type:    "io.genesisdb.app.user-created",
        Data: map[string]interface{}{
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
        },
    },
}

preconditions := []genesisdb.Precondition{
    {
        Type: "isSubjectNew",
        Payload: map[string]interface{}{
            "subject": "/user/456",
        },
    },
}

err := client.CommitEventsWithPreconditions(events, preconditions)
if err != nil {
    log.Fatal(err)
}
```

### isQueryResultTrue
Evaluates a query and ensures the result is truthy. Supports the full GDBQL feature set including complex WHERE clauses, aggregations, and calculated fields.

**Basic uniqueness check:**
```go
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.app",
        Subject: "/user/456",
        Type:    "io.genesisdb.app.user-created",
        Data: map[string]interface{}{
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
        },
    },
}

preconditions := []genesisdb.Precondition{
    {
        Type: "isQueryResultTrue",
        Payload: map[string]interface{}{
            "query": "STREAM e FROM events WHERE e.data.email == 'john.doe@example.com' MAP COUNT() == 0",
        },
    },
}

err := client.CommitEventsWithPreconditions(events, preconditions)
if err != nil {
    log.Fatal(err)
}
```

**Business rule enforcement (transaction limits):**
```go
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.banking",
        Subject: "/user/123/transactions",
        Type:    "io.genesisdb.banking.transaction-processed",
        Data: map[string]interface{}{
            "amount": 500.00,
            "currency": "EUR",
        },
    },
}

preconditions := []genesisdb.Precondition{
    {
        Type: "isQueryResultTrue",
        Payload: map[string]interface{}{
            "query": "STREAM e FROM events WHERE e.subject UNDER '/user/123' AND e.type == 'transaction-processed' AND e.time >= '2024-01-01T00:00:00Z' MAP SUM(e.data.amount) + 500 <= 10000",
        },
    },
}

err := client.CommitEventsWithPreconditions(events, preconditions)
if err != nil {
    log.Fatal(err)
}
```

**Complex validation with aggregations:**
```go
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.events",
        Subject: "/conference/2024/registrations",
        Type:    "io.genesisdb.events.registration-created",
        Data: map[string]interface{}{
            "attendeeId": "att-789",
            "ticketType": "premium",
        },
    },
}

preconditions := []genesisdb.Precondition{
    {
        Type: "isQueryResultTrue",
        Payload: map[string]interface{}{
            "query": "STREAM e FROM events WHERE e.subject UNDER '/conference/2024/registrations' AND e.type == 'registration-created' GROUP BY e.data.ticketType HAVING e.data.ticketType == 'premium' MAP COUNT() < 50",
        },
    },
}

err := client.CommitEventsWithPreconditions(events, preconditions)
if err != nil {
    log.Fatal(err)
}
```

**Supported GDBQL Features in Preconditions:**
- WHERE conditions with AND/OR/IN/BETWEEN operators
- Hierarchical subject queries (UNDER, DESCENDANTS)
- Aggregation functions (COUNT, SUM, AVG, MIN, MAX)
- GROUP BY with HAVING clauses
- ORDER BY and LIMIT clauses
- Calculated fields and expressions
- Nested field access (e.data.address.city)
- String concatenation and arithmetic operations

If a precondition fails, the commit returns HTTP 412 (Precondition Failed) with details about which condition failed.

### Querying Events

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

query := `
STREAM e FROM events
WHERE e.type == 'io.genesisdb.app.customer-added'
ORDER BY e.time
MAP { id: e.id, firstName: e.data.firstName, lastName: e.data.lastName }
`

results, err := client.Q(query)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Result: %v\n", result)
}
```

### Querying Events (Alternative Method)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

query := `STREAM e FROM events WHERE e.type == "io.genesisdb.app.customer-added" ORDER BY e.time DESC LIMIT 20 MAP { subject: e.subject, firstName: e.data.firstName }`

results, err := client.QueryEvents(query)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Result: %v\n", result)
}
```


## Health Checks

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

// Ping the API
response, err := client.Ping()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Ping response: %s\n", response)

// Run audit to check event consistency
response, err = client.Audit()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Audit response: %s\n", response)
```

## Error Handling

All methods return errors when something goes wrong. Make sure to check for errors and handle them appropriately.

## License

MIT

## Author

* E-Mail: mail@genesisdb.io
* URL: https://www.genesisdb.io
* Docs: https://docs.genesisdb.io
