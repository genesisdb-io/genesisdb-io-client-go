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

Copyright © 2024 Genesis DB. All rights reserved.
