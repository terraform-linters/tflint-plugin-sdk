## 0.5.0 (2020-09-13)

This release adds `Config()` API to accessing the Terraform configuration. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.20+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/15 for an example of upgrading the SDK. 

### Breaking Changes

- [#59](https://github.com/terraform-linters/tflint-plugin-sdk/pull/59): Make terraform configuration compatible with v0.13
  - Remove `Backend.ConfigRange`, `ModleCall.ConfigRange`, `ModuleCall.CountRange`, and `ModuleCall.ForEachRange`
- [#65](https://github.com/terraform-linters/tflint-plugin-sdk/pull/65): Tweaks package structures
  - Renamed package names
    - `terraform.Backend` => `configs.Backend`
    - `terraform.Resource` => `configs.Resource`
    - `terraform.ModuleCall` => `configs.ModuleCall`
- [#67](https://github.com/terraform-linters/tflint-plugin-sdk/pull/67): plugin: Bump protocol version

### Enhancements

- [#60](https://github.com/terraform-linters/tflint-plugin-sdk/pull/60): tflint: Add `Runner.Config()`

### Internal Changes

- [#54](https://github.com/terraform-linters/tflint-plugin-sdk/pull/54): Accept DisabledByDefault config attribute

### Chores

- [#55](https://github.com/terraform-linters/tflint-plugin-sdk/pull/55): chore(deps): bump go to v1.15
- [#56](https://github.com/terraform-linters/tflint-plugin-sdk/pull/56): Update GitHub Actions by Dependabot
- [#57](https://github.com/terraform-linters/tflint-plugin-sdk/pull/57): Bump actions/setup-go from v1 to v2.1.2
- [#58](https://github.com/terraform-linters/tflint-plugin-sdk/pull/58): Bump github.com/google/go-cmp from 0.5.1 to 0.5.2
- [#63](https://github.com/terraform-linters/tflint-plugin-sdk/pull/63): Bump github.com/zclconf/go-cty from 1.5.1 to 1.6.1
- [#66](https://github.com/terraform-linters/tflint-plugin-sdk/pull/66): Make terraform configuration compatible with v0.13.2

## 0.4.0 (2020-08-17)

This release adds `WalkModuleCalls()` API to accessing the module calls. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.19+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/11 for an example of upgrading the SDK. 

### Breaking Changes

- [#53](https://github.com/terraform-linters/tflint-plugin-sdk/pull/53): plugin: Bump protocol version

### Enhancements

- [#50](https://github.com/terraform-linters/tflint-plugin-sdk/pull/50): client: Implement `WalkModuleCalls`

### Chores

- [#51](https://github.com/terraform-linters/tflint-plugin-sdk/pull/51): Bump github.com/google/go-cmp from 0.5.0 to 0.5.1
- [#52](https://github.com/terraform-linters/tflint-plugin-sdk/pull/52): Bump github.com/hashicorp/go-version from 1.0.0 to 1.2.1

## 0.3.0 (2020-07-19)

This release adds `Backend()` API to accessing the Terraform backend configuration. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.18+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/10 for an example of upgrading the SDK. 

### Breaking Changes

- [#48](https://github.com/terraform-linters/tflint-plugin-sdk/pull/48): plugin: Bump protocol version

### Enhancements

- [#47](https://github.com/terraform-linters/tflint-plugin-sdk/pull/47): client: Add `Runner.Backend()`
- [#49](https://github.com/terraform-linters/tflint-plugin-sdk/pull/49): helper: Add Backend() helper

## 0.2.0 (2020-06-27)

This release adds APIs to access more Terraform's configurations.

Previously, only `WalkResourceAttributes` that can access top-level attributes were available, but `WalkResourceBlocks` that can access blocks and `WalkResources` that can access the entire resource including meta-arguments are now available.

In addition, the communication system between the plugin and the host has changed, and it is no longer dependent on the HCL structure implementation. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.17+.

### Breaking Changes

- [#24](https://github.com/terraform-linters/tflint-plugin-sdk/pull/24): tflint: Sending expression nodes as a text representation
- [#41](https://github.com/terraform-linters/tflint-plugin-sdk/pull/41): tflint: Remove Metadata
- [#45](https://github.com/terraform-linters/tflint-plugin-sdk/pull/45): plugin: Bump protocol version

### Enhancements

- [#23](https://github.com/terraform-linters/tflint-plugin-sdk/pull/23): helper: Compare Rule types with custom comparer
- [#29](https://github.com/terraform-linters/tflint-plugin-sdk/pull/29) [#33](https://github.com/terraform-linters/tflint-plugin-sdk/pull/33): tflint: Add WalkResourceBlocks API
- [#35](https://github.com/terraform-linters/tflint-plugin-sdk/pull/35): tflint: Allow to omit metadata expr on EmitIssue
- [#34](https://github.com/terraform-linters/tflint-plugin-sdk/pull/34) [#37](https://github.com/terraform-linters/tflint-plugin-sdk/pull/37): tflint: Add WalkResources API
- [#40](https://github.com/terraform-linters/tflint-plugin-sdk/pull/40): helper: Add WalkResourceBlocks helper

### Chores

- [#27](https://github.com/terraform-linters/tflint-plugin-sdk/pull/27): Bump github.com/hashicorp/go-hclog from 0.13.0 to 0.14.1
- [#38](https://github.com/terraform-linters/tflint-plugin-sdk/pull/38): Revise package structure
- [#39](https://github.com/terraform-linters/tflint-plugin-sdk/pull/39): Bump github.com/google/go-cmp from 0.4.1 to 0.5.0
- [#42](https://github.com/terraform-linters/tflint-plugin-sdk/pull/42): Bump github.com/zclconf/go-cty from 1.4.1 to 1.5.1
- [#43](https://github.com/terraform-linters/tflint-plugin-sdk/pull/43): Create Dependabot config file
- [#44](https://github.com/terraform-linters/tflint-plugin-sdk/pull/44): Setup Code Scanning

## 0.1.1 (2020-05-23)

### Changes

- [#20](https://github.com/terraform-linters/tflint-plugin-sdk/pull/20): helper: Make Issues of TestRunner non-nil

### Chores

- [#15](https://github.com/terraform-linters/tflint-plugin-sdk/pull/15): Bump github.com/hashicorp/go-hclog from 0.10.1 to 0.13.0
- [#16](https://github.com/terraform-linters/tflint-plugin-sdk/pull/16): Bump github.com/google/go-cmp from 0.3.1 to 0.4.1
- [#17](https://github.com/terraform-linters/tflint-plugin-sdk/pull/17): Bump github.com/hashicorp/hcl/v2 from 2.2.0 to 2.5.1
- [#18](https://github.com/terraform-linters/tflint-plugin-sdk/pull/18): Adding gitignore file
- [#19](https://github.com/terraform-linters/tflint-plugin-sdk/pull/19): Add tests and CI setups
- [#21](https://github.com/terraform-linters/tflint-plugin-sdk/pull/21): Bump github.com/hashicorp/go-plugin from 1.0.1 to 1.3.0
- [#22](https://github.com/terraform-linters/tflint-plugin-sdk/pull/22): Bump github.com/zclconf/go-cty from 1.1.1 to 1.4.1

## 0.1.0 (2020-01-18)

Initial release
