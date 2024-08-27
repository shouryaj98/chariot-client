# Chariot SDK

The SDK is a Go package that provides a programmatic interface to the Chariot API. 

## Setup

The SDK requires the instantiation of a `keychain.ini` file describing the target backend and credentials to access.

The base keychain.ini file can be downloaded from `preview.chariot.praetorian.com/keychain.ini` and will look something like this (note the addition of the username + password fields):

```ini
[United States]
name = chariot
client_id = 795dnnr45so7m17cppta0b295o
api = https://d0qcl2e18h.execute-api.us-east-2.amazonaws.com/chariot
user_pool_id = us-east-2_BJ6QHVG2L
username = <your username here>
password = <your password here>
```

This file should be saved to `$HOME/.chariot/keychain.ini`.

## Example Usage

Once your `keychain.ini` file is setup, you can use the SDK to interact with the Chariot API. Here's an example of how to list all assets in the system:

```go
package main

import (
    "fmt"
    "github.com/praetorian-inc/chariot-client/pkg/sdk"
)

func main() {
	// Note that the client will use whatever profile is specified in the call to NewClient, but an empty string will
	// default to reading the first profile at the top of your keychain.ini file.
    client, err := sdk.NewClient("")
    if err != nil {
        fmt.Println(err)
        return
    }
	
	assets, err := Client.Assets.List()
	if err != nil {
		cmd.Printf("Failed to list assets: %v\n", err)
		return
	}

    for _, asset := range assets {
        fmt.Println(asset.Key)
    }
}
```

More examples of how to use the SDK can be found in the `./internal/commands` directory which uses the SDK to implement our CLI.