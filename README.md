# TFLint plugin SDK
[![Build Status](https://github.com/terraform-linters/tflint-plugin-sdk/workflows/build/badge.svg?branch=master)](https://github.com/terraform-linters/tflint-plugin-sdk/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint-plugin-sdk.svg)](https://github.com/terraform-linters/tflint-plugin-sdk/releases/latest)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

[TFLint](https://github.com/terraform-linters/tflint) plugin SDK for building custom rules.

NOTE: This plugin system is experimental. This means that API compatibility is frequently broken.

## Requirements

- TFLint v0.40+
- Go v1.20

## Usage

Please refer to [tflint-ruleset-template](https://github.com/terraform-linters/tflint-ruleset-template) for an example plugin implementation using this SDK.

For more details on the API, see [tflint](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk/tflint) and [helper](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk/helper) packages on pkg.go.dev.

## Developing

The proto compiler is required when updating `.proto` files. The `protoc` and `protoc-gen-go` can be installed using [aqua](https://github.com/aquaproj/aqua).

```console
$ make prepare
curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v2.1.1/aqua-installer | bash
===> Installing aqua v2.2.3 for bootstraping...

...

aqua version 2.3.7 (c07105b10ab825e7f309d2eb83278a0422a2b24f)

Add ${AQUA_ROOT_DIR}/bin to the environment variable PATH.
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
$ export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
$ aqua install
$ make proto
```

## Architecture

![architecture](architecture.png)

This plugin system uses [go-plugin](https://github.com/hashicorp/go-plugin). TFLint launches the plugin as a sub-process and communicates with the plugin over gRPC. The plugin acts as a server, while TFLint acts as a client that sends inspection requests to the plugin.

On the other hand, the plugin sends various requests to a server (TFLint) to get detailed runtime contexts (e.g. variables and expressions). This means that TFLint and plugins can act as both a server and a client.

These implementations are included in the [plugin/host2plugin](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk/plugin/host2plugin) and [plugin/plugin2host](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host) packages.
