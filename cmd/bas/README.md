# Chariot BAS

What makes a good TTP? Code that executes a known adversarial behavior that should be prevented in a secure environment.

## Building

Run `make bas` from the base of the `chariot-client` repository to build all BAS TTPs.

# Usage

To write a TTP, create a ``.go`` file in the ``PROJECT_ROOT/cmd/bas`` directory, using the template below.

```go
package tests

import "github.com/praetorian-inc/chariot-client/pkg/bas/endpoint"

func test() {
    // STOP with a predefined condition
    // review codes.go for all options
    endpoint.Stop(endpoint.Risk.Allowed)
}

func cleanup() {
    // optional logic to reverse the effects of this test
}

func main(){
    endpoint.Start(test, cleanup)
}
```

Compile your new TTP from the base directory of `chariot-client`:

```sh
guid="<UUID>"

echo "// $(date)" >> ./cmd/bas/${guid}.go
make bas
```

For example for `8d4db53ba95e96524a3413518193a1b4`:

```
GOOS=windows GOARCH=amd64 go build -o 8d4db53ba95e96524a3413518193a1b4.exe 8d4db53ba95e96524a3413518193a1b4.go
```
