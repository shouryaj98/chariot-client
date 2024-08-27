# Chariot Client

The client is a do-it-all binary that enables easy access to the Chariot offensive security platform.

## Building

Run `make` to build all targets in this repository. This will build:

- A command-line interface for programmatic API access
- A Docker image for running standalone capabilities (code that either discovers assets or identifies risks on them)
- A collection of Breach and Attack Simulation (BAS) TTPs

## Features

### CLI

The command-line interface provides complete access to the Chariot API. It is built on top of a Go SDK and is the programmatic sibling to the [Chariot UI](https://github.com/praetorian-inc/chariot-ui).

To build the cli, run:

```
make cli
```

This will place a binary called `chariot` in the `./build/bin` directory. Feel free to add this to your `$PATH`.

### ASM

A capability is a piece of code that either discovers an asset or identifies a risk on an asset. Chariot ASM contains capabilities that are imported by the closed-source SaaS platform (https://preview.chariot.praetorian.com) but are also standalone and entirely open-source.

[Build your own capability](./pkg/asm/capabilities)

To build the docker image containing all supported capabilities, run:

```
make asm
```

this will build the container with the name `chariot-asm`

### BAS

Breach and Attack Simulation represents a collection of MITRE ATT&CK TTPs that test endpoint defenses against malicious behaviors. Chariot TTPs can be compiled and run standalone - or uploaded to the SaaS platform for distributed testing.

[Build your own TTP](./cmd/bas)

All BAS TTPs can be built using the command

```
make bas
```

The resulting binaries will be in `./build/bin` and will be compiled for Windows by default.

### SDK

`chariot-client` also houses the public [Chariot golang SDK](./pkg/sdk). The code here is designed to be imported into other go progams that need to interact with the Chariot public API.
