# Go CleverCloud API client

[![Go Reference](https://pkg.go.dev/badge/go.clever-cloud.dev/client.svg)](https://pkg.go.dev/go.clever-cloud.dev/client)

## How to

### Installation
Add this client as project dependency
```sh
go get -u go.clever-cloud.dev/client
```

### Usage

#### Client instantiation

```go
import "go.clever-cloud.dev/client"

cc := client.New(
    client.WithAutoOauthConfig(),
) 

```

#### Use the client

```go
type Self struct {
    ID string `json:"id"`
}

res := client.Get[Self](context.Background(), cc, "/v2/self")
if res.HasError() {
    // handle res.Error()
}

fmt.Println(res.Payload().ID)

```

if the operation you want to do does not return anything, use:
```go
res := client.Get[client.Nothing](context.Background(), cc, "/v2/self")
```