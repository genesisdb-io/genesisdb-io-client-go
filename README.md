# Genesis DB Go SDK

This is the official Go SDK for Genesis DB. It provides a simple interface to interact with the Genesis DB API.

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

### Stream Events with latest by event type

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

options := &genesisdb.StreamOptions{
    LatestByEventType: "io.genesisdb.foo.foobarfoo-updated",
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
        Subject: "/foo/21",
        Type:    "io.genesisdb.app.foo-added",
        Data: map[string]interface{}{
            "value": "Foo",
        },
        Options: map[string]interface{}{
            "storeDataAsReference": true,
        },
    },
}

err := client.CommitEvents(events)
if err != nil {
    log.Fatal(err)
}
```

### Deleting referenced data (GDPR)

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

err := client.EraseData("/foo/21")
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
        Subject: "/foo/21",
        Type:    "io.genesisdb.app.foo-added",
        Data: map[string]interface{}{
            "value": "Foo",
        },
    },
}

preconditions := []genesisdb.Precondition{
    {
        Type: "isSubjectNew",
        Payload: map[string]interface{}{
            "subject": "/foo/21",
        },
    },
}

err := client.CommitEventsWithPreconditions(events, preconditions)
if err != nil {
    log.Fatal(err)
}
```

### isQueryResultTrue
Evaluates a query and ensures the result is truthy:

```go
events := []genesisdb.Event{
    {
        Source: "io.genesisdb.app",
        Subject: "/event/conf-2024",
        Type:    "io.genesisdb.app.registration-added",
        Data: map[string]interface{}{
            "attendeeName": "Alice",
            "eventId": "conf-2024",
        },
    },
}

preconditions := []genesisdb.Precondition{
    {
        Type: "isQueryResultTrue",
        Payload: map[string]interface{}{
            "query": "FROM e IN events WHERE e.data.eventId == 'conf-2024' PROJECT INTO COUNT() < 500",
        },
    },
}

err := client.CommitEventsWithPreconditions(events, preconditions)
if err != nil {
    log.Fatal(err)
}
```

### Querying Events

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

query := `
FROM e IN events
WHERE e.type == 'io.genesisdb.app.customer-added'
ORDER BY e.time
PROJECT INTO { id: e.id, firstName: e.data.firstName, lastName: e.data.lastName }
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

query := `FROM e IN events WHERE e.type == "io.genesisdb.app.customer-added" ORDER BY e.time DESC TOP 20 PROJECT INTO { subject: e.subject, firstName: e.data.firstName }`

results, err := client.QueryEvents(query)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Result: %v\n", result)
}
```

### Health Checks

```go
import "github.com/genesisdb-io/genesisdb-io-client-go/pkg/genesisdb"

// Ping the API
response, err := client.Ping()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Ping response: %s\n", response)

// Run audit
response, err = client.Audit()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Audit response: %s\n", response)
```

## Error Handling

All methods return errors when something goes wrong. Make sure to check for errors and handle them appropriately.

## License

Copyright © 2025 Genesis DB. All rights reserved.
