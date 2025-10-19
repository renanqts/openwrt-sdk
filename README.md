# OpenWRT SDK
A Go library for interacting with OpenWRT routers via the LuCI RPC API.

## Features
- LuCI RPC API integration
- SDK to interacte with the Router

## Installation
```
go get github.com/renanqts/openwrt-sdk
```

## Usage
Here's a basic example of how to use the SDK:

```go
package main

import (
    "context"
    "log"
    
    "github.com/renanqts/openwrt-sdk/pkg/sdk"
)

func main() {
    // Create a new OpenWRT client
    client, err := sdk.New(
        "https://192.168.1.1",  // Router address
        "admin",                // Username
        "password",             // Password
        1,                      // RPC ID
        false,                  // Skip TLS verification
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use the client to interact with the router
}
```


## Development
Requirements
- Go 1.24.2 or higher
- golangci-lint for code quality checks

### Testing
The project includes mock implementations generated using mockgen.
To run tests:
```sh
go test ./...
```

To generate mocks:
```sh
go install go.uber.org/mock/mockgen@latest
go test ./...
```

## Contributing
1. Fork the repository
1. Create your feature branch (`git checkout -b feature/amazing-feature`)
1. Commit your changes (`git commit -m 'Add some amazing feature'`)
1. Push to the branch (`git push origin feature/amazing-feature`)
1. Open a Pull Request

## Acknowledgments
- OpenWRT project for their excellent router firmware
- LuCI team for the RPC API

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/renanqts4)