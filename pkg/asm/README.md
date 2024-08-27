# Chariot ASM

A capability is a program that automates common actions to identify assets and risks.

## Building

Run `make asm` from the base of the `chariot-client` repository to build a TTPs.

## Usage

Start the Docker container (we call them compute servers):

```sh
make compute -C ../..
```

Run a capability using:

``docker run -it --rm chariot-asm --capability <name> --name example.com``

## Contributing

To write a new capability, create a ``.go`` file in the ``PROJECT_ROOT/pkg/asm/capabilities`` directory, using the template below that creates a new asset and new risk with a proof of exploit for that asset. 

To hook the capability into the system, update the `Registry` map in ``PROJECT_ROOT/cmd/asm/main.go`` with an additional entry for  `"dummy": capabilities.NewDummy,`.

```go
package capabilities

import "github.com/praetorian-inc/chariot-client/pkg/sdk/model"

type Dummy struct {
	Job   model.Job
	Asset model.Asset
	XYZ
}

func NewDummy(job model.Job) model.Capability {
	return &Dummy{Asset: job.Target, Job: job, XYZ: NewXYZ()}
}

func (task *Dummy) Match() bool {
	return task.Asset.Is("domain")
}

func (task *Dummy) Invoke() error {
	identified := model.NewAsset("example.com", "example.com")
	risk := model.NewRisk(identified, "example-risk")

	task.Job.Stream <- identified
	task.Job.Stream <- risk
	task.Job.Stream <- risk.Proof([]byte("example proof of exploit"))

	return nil
}
```


