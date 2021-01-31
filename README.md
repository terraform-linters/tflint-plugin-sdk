# TFLint plugin SDK
[![Build Status](https://github.com/terraform-linters/tflint-plugin-sdk/workflows/build/badge.svg?branch=master)](https://github.com/terraform-linters/tflint-plugin-sdk/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint-plugin-sdk.svg)](https://github.com/terraform-linters/tflint-plugin-sdk/releases/latest)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

[TFLint](https://github.com/terraform-linters/tflint) plugin SDK for building custom rules.

NOTE: This plugin system is experimental. This means that API compatibility is frequently broken.

## Requirements

- TFLint v0.24+
- Go v1.15

## Usage

Please refer to [tflint-ruleset-template](https://github.com/terraform-linters/tflint-ruleset-template) for an example plugin implementation using this SDK.

For more details on the API, see [tflint](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk/tflint) and [helper](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk/helper) packages on pkg.go.dev.

## Architecture

![architecture](architecture.png)

This plugin system uses [go-plugin](https://github.com/hashicorp/go-plugin). TFLint launches the plugin as a sub-process and communicates with the plugin over RPC. The plugin acts as a server, while TFLint acts as a client that sends inspection requests to the plugin.

On the other hand, the plugin sends various requests to a server (TFLint) to get detailed runtime contexts (e.g. variables and expressions). This means that TFLint and plugins can act as both a server and a client.
