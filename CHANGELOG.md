## 0.20.0 (2024-05-02)

### Enhancements

- [#316](https://github.com/terraform-linters/tflint-plugin-sdk/pull/316): Bump github.com/hashicorp/hcl/v2 from 2.19.1 to 2.20.1
  - This is required for provider-defined functions support

### Chores

- [#311](https://github.com/terraform-linters/tflint-plugin-sdk/pull/311): Bump google.golang.org/protobuf from 1.32.0 to 1.33.0
- [#314](https://github.com/terraform-linters/tflint-plugin-sdk/pull/314): Bump github.com/zclconf/go-cty from 1.14.2 to 1.14.4
- [#317](https://github.com/terraform-linters/tflint-plugin-sdk/pull/317): Bump github.com/hashicorp/go-hclog from 1.6.2 to 1.6.3
- [#319](https://github.com/terraform-linters/tflint-plugin-sdk/pull/319): Bump golang.org/x/tools from 0.18.0 to 0.20.0
- [#320](https://github.com/terraform-linters/tflint-plugin-sdk/pull/320): Bump google.golang.org/grpc from 1.61.1 to 1.63.2
- [#321](https://github.com/terraform-linters/tflint-plugin-sdk/pull/321): Bump golang.org/x/net from 0.21.0 to 0.23.0
- [#322](https://github.com/terraform-linters/tflint-plugin-sdk/pull/322): deps: Go 1.22.2
- [#323](https://github.com/terraform-linters/tflint-plugin-sdk/pull/323): Bump google.golang.org/protobuf from 1.33.0 to 1.34.0

## 0.19.0 (2024-02-25)

### Chores

- [#281](https://github.com/terraform-linters/tflint-plugin-sdk/pull/281) [#307](https://github.com/terraform-linters/tflint-plugin-sdk/pull/307): deps: Go 1.22
- [#283](https://github.com/terraform-linters/tflint-plugin-sdk/pull/283): Bump actions/checkout from 3 to 4
- [#294](https://github.com/terraform-linters/tflint-plugin-sdk/pull/294): Bump github.com/google/go-cmp from 0.5.9 to 0.6.0
- [#297](https://github.com/terraform-linters/tflint-plugin-sdk/pull/297): Bump github.com/hashicorp/hcl/v2 from 2.17.0 to 2.19.1
- [#299](https://github.com/terraform-linters/tflint-plugin-sdk/pull/299): Bump google.golang.org/grpc from 1.57.0 to 1.61.1
- [#300](https://github.com/terraform-linters/tflint-plugin-sdk/pull/300): Bump github.com/hashicorp/go-plugin from 1.4.10 to 1.6.0
- [#301](https://github.com/terraform-linters/tflint-plugin-sdk/pull/301): Bump github.com/zclconf/go-cty from 1.13.2 to 1.14.2
- [#302](https://github.com/terraform-linters/tflint-plugin-sdk/pull/302): Bump golang.org/x/tools from 0.11.0 to 0.18.0
- [#303](https://github.com/terraform-linters/tflint-plugin-sdk/pull/303): Bump github/codeql-action from 2 to 3
- [#304](https://github.com/terraform-linters/tflint-plugin-sdk/pull/304): Bump actions/setup-go from 4 to 5
- [#305](https://github.com/terraform-linters/tflint-plugin-sdk/pull/305): Bump github.com/hashicorp/go-hclog from 1.5.0 to 1.6.2
- [#306](https://github.com/terraform-linters/tflint-plugin-sdk/pull/306): Bump google.golang.org/protobuf from 1.31.0 to 1.32.0

## 0.18.0 (2023-07-29)

### Breaking Changes

- [#268](https://github.com/terraform-linters/tflint-plugin-sdk/pull/268): helper: Un-export `NewLocalRunner` and `AddLocalFile`
- [#273](https://github.com/terraform-linters/tflint-plugin-sdk/pull/273): Internalize unnecessarily published plugin packages
- [#274](https://github.com/terraform-linters/tflint-plugin-sdk/pull/274): tflint: Remove deprecated IncludeNotCreated option

### Enhancements

- [#271](https://github.com/terraform-linters/tflint-plugin-sdk/pull/271): hclext: Add support for expression unwrapping in hclext.BoundExpr

### Chores

- [#267](https://github.com/terraform-linters/tflint-plugin-sdk/pull/267): Bump google.golang.org/protobuf from 1.30.0 to 1.31.0
- [#270](https://github.com/terraform-linters/tflint-plugin-sdk/pull/270): Bump golang.org/x/tools from 0.10.0 to 0.11.0
- [#272](https://github.com/terraform-linters/tflint-plugin-sdk/pull/272): Bump google.golang.org/grpc from 1.55.0 to 1.57.0

## 0.17.0 (2023-06-18)

This release adds support for autofix API. The `EmitIssueWithFix` API allows you to implement autofix in your plugin using `tflint.Fixer`. Autofix is available in TFLint v0.47+. In earlier versions, the autofix is ignored.

This SDK version no longer supports TFLint v0.40/v0.41. This means that plugins built using this SDK require TFLint v0.42+.

Also, the `Check` method has been removed from `tflint.RuleSet` as a minor change. This means that if you override the `Check` method in a custom ruleset that embeds `tflint.RuleSet`, it will not be called. This is classified as a breaking change, but since the `Check` method is not supposed to be overwritten, it is recommended to use something like `NewRunner`.

### Breaking Changes

- [#258](https://github.com/terraform-linters/tflint-plugin-sdk/pull/258): tflint: Remove `Check` method from `tflint.RuleSet` interface
- [#263](https://github.com/terraform-linters/tflint-plugin-sdk/pull/263): Drop support for TFLint v0.40/v0.41

### Enhancements

- [#254](https://github.com/terraform-linters/tflint-plugin-sdk/pull/254): Introduce autofix API

### Chores

- [#253](https://github.com/terraform-linters/tflint-plugin-sdk/pull/253): Configure aqua to install protoc
- [#255](https://github.com/terraform-linters/tflint-plugin-sdk/pull/255): Bump google.golang.org/grpc from 1.54.0 to 1.55.0
- [#257](https://github.com/terraform-linters/tflint-plugin-sdk/pull/257): Bump github.com/zclconf/go-cty from 1.13.1 to 1.13.2
- [#261](https://github.com/terraform-linters/tflint-plugin-sdk/pull/261): Bump github.com/hashicorp/go-plugin from 1.4.9 to 1.4.10
- [#262](https://github.com/terraform-linters/tflint-plugin-sdk/pull/262): Bump github.com/hashicorp/hcl/v2 from 2.16.2 to 2.17.0
- [#264](https://github.com/terraform-linters/tflint-plugin-sdk/pull/264): Bump golang.org/x/tools from 0.8.0 to 0.10.0

## 0.16.1 (2023-04-13)

### BugFixes

- [#252](https://github.com/terraform-linters/tflint-plugin-sdk/pull/252): ruleset: Fix NewRunner hook not injecting a custom runner

### Chores

- [#251](https://github.com/terraform-linters/tflint-plugin-sdk/pull/251): Bump golang.org/x/tools from 0.7.0 to 0.8.0

## 0.16.0 (2023-04-02)

This release deprecates the `runner.EnsureNoError` helper. This helper is still available in this version, but we recommend migrating to the function callback approach.

```go
// Before
var val string
err := runner.EvaluateExpr(expr, &val, nil)
err = runner.EnsureNoError(err, func () error {
  // Test values
})
if err != nil {
  return err
}

// After
err := runner.EvaluateExpr(expr, func (val string), error {
  // Test values
}, nil)
```

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/76 for an example of upgrading the SDK.

### Enhancements

- [#225](https://github.com/terraform-linters/tflint-plugin-sdk/pull/225): ruleset: Allow a runner to be redefined within a ruleset
  - The `NewRunner` method has been added to the `tflint.RuleSet` interface.
- [#239](https://github.com/terraform-linters/tflint-plugin-sdk/pull/239): plugin2host: Send marked values over the wire
  - With this change, sensitive values can now be handled by plugins (requires TFLint v0.46+). Previously, `tflint.ErrSensitive` was always returned.
- [#246](https://github.com/terraform-linters/tflint-plugin-sdk/pull/246) [#247](https://github.com/terraform-linters/tflint-plugin-sdk/pull/247): runner: Add support for function callbacks as the target of `EvaluateExpr`
  - This allows reproducing the same behavior as before without using `EnsureNoError`.
- [#248](https://github.com/terraform-linters/tflint-plugin-sdk/pull/248): runner: Add support for the bool type as a target value of `EvaluateExpr`

### Changes

- [#236](https://github.com/terraform-linters/tflint-plugin-sdk/pull/236): runner: Deprecate `EnsureNoError` helper
  - This helper is still available in this version, but we recommend migrating to the function callback approach.

### Chores

- [#233](https://github.com/terraform-linters/tflint-plugin-sdk/pull/233): Bump golang.org/x/net from 0.3.0 to 0.7.0
- [#234](https://github.com/terraform-linters/tflint-plugin-sdk/pull/234): Go 1.20
- [#235](https://github.com/terraform-linters/tflint-plugin-sdk/pull/235): plugin2host: Handle eval errors on the client side
- [#238](https://github.com/terraform-linters/tflint-plugin-sdk/pull/238): Bump github.com/hashicorp/go-plugin from 1.4.8 to 1.4.9
- [#240](https://github.com/terraform-linters/tflint-plugin-sdk/pull/240): Bump github.com/hashicorp/hcl/v2 from 2.15.0 to 2.16.2
- [#241](https://github.com/terraform-linters/tflint-plugin-sdk/pull/241): Bump golang.org/x/tools from 0.4.0 to 0.7.0
- [#243](https://github.com/terraform-linters/tflint-plugin-sdk/pull/243): Bump actions/setup-go from 3 to 4
- [#244](https://github.com/terraform-linters/tflint-plugin-sdk/pull/244): Bump github.com/zclconf/go-cty from 1.12.1 to 1.13.1
- [#245](https://github.com/terraform-linters/tflint-plugin-sdk/pull/245): Bump google.golang.org/protobuf from 1.28.1 to 1.30.0
- [#249](https://github.com/terraform-linters/tflint-plugin-sdk/pull/249): Bump github.com/hashicorp/go-hclog from 1.4.0 to 1.5.0
- [#250](https://github.com/terraform-linters/tflint-plugin-sdk/pull/250): Bump google.golang.org/grpc from 1.51.0 to 1.54.0

## 0.15.0 (2022-12-26)

### Enhancements

- [#224](https://github.com/terraform-linters/tflint-plugin-sdk/pull/224): Add GetOriginalwd method

### Chores

- [#214](https://github.com/terraform-linters/tflint-plugin-sdk/pull/214): Bump github.com/hashicorp/hcl/v2 from 2.14.1 to 2.15.0
- [#219](https://github.com/terraform-linters/tflint-plugin-sdk/pull/219): Bump google.golang.org/grpc from 1.50.1 to 1.51.0
- [#220](https://github.com/terraform-linters/tflint-plugin-sdk/pull/220): Bump github.com/hashicorp/go-plugin from 1.4.5 to 1.4.8
- [#221](https://github.com/terraform-linters/tflint-plugin-sdk/pull/221): Bump github.com/go-test/deep from 1.0.8 to 1.1.0
- [#222](https://github.com/terraform-linters/tflint-plugin-sdk/pull/222): Bump github.com/hashicorp/go-hclog from 1.3.1 to 1.4.0
- [#223](https://github.com/terraform-linters/tflint-plugin-sdk/pull/223): Bump golang.org/x/tools from 0.1.12 to 0.4.0

## 0.14.0 (2022-10-23)

This release includes several new features for plugin developers. Introduced the Schema Mode to get all attributes, and added an option to set constraints on compatible TFLint versions. These may not work with older TFLint versions, so set version constraints as needed.

The evaluation of `each.*` and `count.*` added in TFLint v0.42 requires plugins built with this version. In earlier versions, these values are always unknown.

`IncludeNotCreated` in `GetModuleContentOption` has been deprecated. Use `ExpandModeNone` instead. The old option will still work, but will be removed in a future version.

### Enhancements

- [#201](https://github.com/terraform-linters/tflint-plugin-sdk/pull/201): hclext: Add schema mode to BodySchema
  - This is available only for TFLint v0.42+. Schema mode is ignored in earlier versions. Set `>= 0.42.0` as a version constraint if you cannot tolerate being ignored.
- [#202](https://github.com/terraform-linters/tflint-plugin-sdk/pull/202): host2plugin: Allow plugins to set host version constraints
  - This is available only for TFLint v0.42+. Version constraints are ignored in earlier versions. Note that version constraints may not work in v0.40, v0.41.
- [#203](https://github.com/terraform-linters/tflint-plugin-sdk/pull/203): host2plugin: Add SDKVersion
- [#205](https://github.com/terraform-linters/tflint-plugin-sdk/pull/205): hclext: Add hclext.BoundExpr
  - This is necessary due to the evaluation of `each.*` and `count.*` added in TFLint v0.42. Plugins not built with SDK v0.14+ will always evaluate to unknown values.
- [#206](https://github.com/terraform-linters/tflint-plugin-sdk/pull/206): hclext: Add Copy() to structures
- [#207](https://github.com/terraform-linters/tflint-plugin-sdk/pull/207): hclext: Add WalkAttribute to hclext.BodyContent
- [#208](https://github.com/terraform-linters/tflint-plugin-sdk/pull/208): plugin2host: Add ExpandMode to GetModuleContentOption
  - `IncludeNotCreated` is deprecated. Use `ExpandModeNone` instread.

### Chores

- [#199](https://github.com/terraform-linters/tflint-plugin-sdk/pull/199): Bump github.com/hashicorp/hcl/v2 from 2.14.0 to 2.14.1
- [#200](https://github.com/terraform-linters/tflint-plugin-sdk/pull/200): Bump github.com/hashicorp/go-hclog from 1.3.0 to 1.3.1
- [#209](https://github.com/terraform-linters/tflint-plugin-sdk/pull/209): Bump google.golang.org/grpc from 1.49.0 to 1.50.1

## 0.13.0 (2022-09-17)

### Enhancements

- [#198](https://github.com/terraform-linters/tflint-plugin-sdk/pull/198): host2plugin: Allow ruleset to accept Only option
  - This change is necessary due to a priority bug with the `--only` option. Most plugins are unaffected by this change.

### Chores

- [#197](https://github.com/terraform-linters/tflint-plugin-sdk/pull/197): Bump github.com/google/go-cmp from 0.5.8 to 0.5.9

## 0.12.0 (2022-09-07)

This release adds `GetModulePath()` API. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.40+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/62 for an example of upgrading the SDK.

### Breaking Changes

- [#171](https://github.com/terraform-linters/tflint-plugin-sdk/pull/171): Add GetModulePath method
- [#188](https://github.com/terraform-linters/tflint-plugin-sdk/pull/188): Bump protocol version

### Enhancements

- [#169](https://github.com/terraform-linters/tflint-plugin-sdk/pull/169): hclext: Add hclext.Blocks's OfType helper
- [#170](https://github.com/terraform-linters/tflint-plugin-sdk/pull/170): hclext: Add AsNative helper
- [#172](https://github.com/terraform-linters/tflint-plugin-sdk/pull/172): tflint: Add GetProviderContent helper
- [#174](https://github.com/terraform-linters/tflint-plugin-sdk/pull/174): tflint: Add tflint.ErrSensitive
- [#177](https://github.com/terraform-linters/tflint-plugin-sdk/pull/177): helper: Add support for JSON syntax in TestRunner
- [#178](https://github.com/terraform-linters/tflint-plugin-sdk/pull/178): Allow calling DecodeRuleConfig without rule config
- [#180](https://github.com/terraform-linters/tflint-plugin-sdk/pull/180): terraform: Add `lang.ReferencesInExpr`
- [#181](https://github.com/terraform-linters/tflint-plugin-sdk/pull/181): tflint: Add WalkExpressions function

### BugFixes

- [#190](https://github.com/terraform-linters/tflint-plugin-sdk/pull/190): logger: Do not set location offset in go-plugin

### Chores

- [#161](https://github.com/terraform-linters/tflint-plugin-sdk/pull/161) [#182](https://github.com/terraform-linters/tflint-plugin-sdk/pull/182): Bump github.com/hashicorp/go-plugin from 1.4.3 to 1.4.5
- [#166](https://github.com/terraform-linters/tflint-plugin-sdk/pull/166) [#194](https://github.com/terraform-linters/tflint-plugin-sdk/pull/194): Bump github.com/hashicorp/hcl/v2 from 2.12.0 to 2.14.0
- [#168](https://github.com/terraform-linters/tflint-plugin-sdk/pull/168) [#187](https://github.com/terraform-linters/tflint-plugin-sdk/pull/187): Bump google.golang.org/grpc from 1.46.0 to 1.49.0
- [#173](https://github.com/terraform-linters/tflint-plugin-sdk/pull/173) [#195](https://github.com/terraform-linters/tflint-plugin-sdk/pull/195): Bump github.com/hashicorp/go-hclog from 1.2.0 to 1.3.0
- [#175](https://github.com/terraform-linters/tflint-plugin-sdk/pull/175): Bump google.golang.org/protobuf from 1.28.0 to 1.28.1
- [#176](https://github.com/terraform-linters/tflint-plugin-sdk/pull/176): build: go 1.19
- [#179](https://github.com/terraform-linters/tflint-plugin-sdk/pull/179): build: Use `go-version-file` instead of `go-version`
- [#183](https://github.com/terraform-linters/tflint-plugin-sdk/pull/183): Bump golang.org/x/tools from 0.1.11 to 0.1.12
- [#184](https://github.com/terraform-linters/tflint-plugin-sdk/pull/184): Bump github.com/go-test/deep from 1.0.3 to 1.0.8
- [#185](https://github.com/terraform-linters/tflint-plugin-sdk/pull/185): Remove unused ruleset function
- [#186](https://github.com/terraform-linters/tflint-plugin-sdk/pull/186): Bump github.com/zclconf/go-cty from 1.10.0 to 1.11.0

## 0.11.0 (2022-05-05)

### Enhancements

- [#160](https://github.com/terraform-linters/tflint-plugin-sdk/pull/160): tflint: Add IncludeNotCreated option to GetModuleContent

### Chores

- [#150](https://github.com/terraform-linters/tflint-plugin-sdk/pull/150): Bump google.golang.org/protobuf from 1.27.1 to 1.28.0
- [#154](https://github.com/terraform-linters/tflint-plugin-sdk/pull/154): Bump actions/setup-go from 2 to 3
- [#155](https://github.com/terraform-linters/tflint-plugin-sdk/pull/155): Bump google.golang.org/grpc from 1.45.0 to 1.46.0
- [#156](https://github.com/terraform-linters/tflint-plugin-sdk/pull/156): Bump github.com/hashicorp/hcl/v2 from 2.11.1 to 2.12.0
- [#157](https://github.com/terraform-linters/tflint-plugin-sdk/pull/157): plugin2host: Return sources instead of `*hcl.File` in GetRuleConfigContent
- [#158](https://github.com/terraform-linters/tflint-plugin-sdk/pull/158): Bump github.com/google/go-cmp from 0.5.7 to 0.5.8
- [#159](https://github.com/terraform-linters/tflint-plugin-sdk/pull/159): Bump github/codeql-action from 1 to 2

## 0.10.1 (2022-04-02)

### BugFixes

- [#153](https://github.com/terraform-linters/tflint-plugin-sdk/pull/153): helper: Skip un-used variable block attributes

## 0.10.0 (2022-03-27)

This release contains a major update to the plugin system. Previously, this SDK uses traditional net/rpc + gob, but now it uses gRPC + Protocol Buffers.

The API also contains many incompatible changes. See https://github.com/terraform-linters/tflint-ruleset-template/pull/48 for how to migrate. TFLint v0.35+ is required to work with new plugin systems.

### Breaking Changes

- [#135](https://github.com/terraform-linters/tflint-plugin-sdk/pull/135) [#146](https://github.com/terraform-linters/tflint-plugin-sdk/pull/146) [#147](https://github.com/terraform-linters/tflint-plugin-sdk/pull/147) [#148](https://github.com/terraform-linters/tflint-plugin-sdk/pull/148) [#149](https://github.com/terraform-linters/tflint-plugin-sdk/pull/149): plugin: gRPC-based new plugin system

### Chores

- [#133](https://github.com/terraform-linters/tflint-plugin-sdk/pull/133) [#145](https://github.com/terraform-linters/tflint-plugin-sdk/pull/145): build: Go 1.18
- [#134](https://github.com/terraform-linters/tflint-plugin-sdk/pull/134): Bump github.com/hashicorp/go-plugin from 1.4.2 to 1.4.3
- [#136](https://github.com/terraform-linters/tflint-plugin-sdk/pull/136): Bump github.com/zclconf/go-cty from 1.9.0 to 1.10.0
- [#138](https://github.com/terraform-linters/tflint-plugin-sdk/pull/138): Bump github.com/hashicorp/hcl/v2 from 2.10.0 to 2.11.1
- [#139](https://github.com/terraform-linters/tflint-plugin-sdk/pull/139): Bump github.com/hashicorp/go-version from 1.3.0 to 1.4.0
- [#141](https://github.com/terraform-linters/tflint-plugin-sdk/pull/141): Bump github.com/google/go-cmp from 0.5.6 to 0.5.7
- [#143](https://github.com/terraform-linters/tflint-plugin-sdk/pull/143): Bump actions/checkout from 2 to 3
- [#144](https://github.com/terraform-linters/tflint-plugin-sdk/pull/144): Bump github.com/hashicorp/go-hclog from 0.16.2 to 1.2.0

## 0.9.1 (2021-07-17)

### BugFixes

- [#128](https://github.com/terraform-linters/tflint-plugin-sdk/pull/128): tflint: Add workaround when parsing a config that has a trailing heredoc

### Chores

- [#125](https://github.com/terraform-linters/tflint-plugin-sdk/pull/125): Bump github.com/zclconf/go-cty from 1.8.4 to 1.9.0
- [#126](https://github.com/terraform-linters/tflint-plugin-sdk/pull/126): Bump github.com/hashicorp/go-hclog from 0.16.1 to 0.16.2

## 0.9.0 (2021-07-03)

This release adds `Files()` API. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.30+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/37 for an example of upgrading the SDK.

### Breaking Changes

- [#122](https://github.com/terraform-linters/tflint-plugin-sdk/pull/122): Implement Files() method
- [#124](https://github.com/terraform-linters/tflint-plugin-sdk/pull/124): Bump protocol version

### Chores

- [#109](https://github.com/terraform-linters/tflint-plugin-sdk/pull/109): Bump github.com/hashicorp/go-version from 1.2.1 to 1.3.0
- [#112](https://github.com/terraform-linters/tflint-plugin-sdk/pull/112): Bump github.com/hashicorp/hcl/v2 from 2.9.1 to 2.10.0
- [#117](https://github.com/terraform-linters/tflint-plugin-sdk/pull/117): Bump github.com/hashicorp/go-hclog from 0.15.0 to 0.16.1
- [#120](https://github.com/terraform-linters/tflint-plugin-sdk/pull/120): Bump github.com/google/go-cmp from 0.5.5 to 0.5.6
- [#121](https://github.com/terraform-linters/tflint-plugin-sdk/pull/121): Bump github.com/hashicorp/go-plugin from 1.4.0 to 1.4.2
- [#123](https://github.com/terraform-linters/tflint-plugin-sdk/pull/123): Bump github.com/zclconf/go-cty from 1.8.1 to 1.8.4

## 0.8.2 (2021-04-04)

### Changes

- [#101](https://github.com/terraform-linters/tflint-plugin-sdk/pull/101): helper: Use a consistent env var for TF_WORKSPACE

### BugFixes

- [#107](https://github.com/terraform-linters/tflint-plugin-sdk/pull/107): client: Pass only type to EvalExpr when passed detailed types

### Chores

- [#102](https://github.com/terraform-linters/tflint-plugin-sdk/pull/102): Upgrade to Go 1.16
- [#103](https://github.com/terraform-linters/tflint-plugin-sdk/pull/103) [#106](https://github.com/terraform-linters/tflint-plugin-sdk/pull/106): Bump github.com/hashicorp/hcl/v2 from 2.8.2 to 2.9.1
- [#105](https://github.com/terraform-linters/tflint-plugin-sdk/pull/105): Bump github.com/google/go-cmp from 0.5.4 to 0.5.5
- [#108](https://github.com/terraform-linters/tflint-plugin-sdk/pull/108): Bump github.com/zclconf/go-cty from 1.8.0 to 1.8.1

## 0.8.1 (2021-02-02)

### BugFixes

- [#100](https://github.com/terraform-linters/tflint-plugin-sdk/pull/100): tflint: Make sure RuleNames always return all rules

## 0.8.0 (2021-01-31)

This release fixes some bugs when using `Config` API. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.24+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/26 for an example of upgrading the SDK. 

### Breaking Changes

- [#96](https://github.com/terraform-linters/tflint-plugin-sdk/pull/96): Use msgpack to encoding to pass cty.Value in variable default
- [#98](https://github.com/terraform-linters/tflint-plugin-sdk/pull/98): Use json for transfering cty.Type

### BugFixes

- [#97](https://github.com/terraform-linters/tflint-plugin-sdk/pull/97): Fix heredoc parsing in parseConfig

### Chores

- [#94](https://github.com/terraform-linters/tflint-plugin-sdk/pull/94): Bump github.com/hashicorp/hcl/v2 from 2.8.1 to 2.8.2
- [#99](https://github.com/terraform-linters/tflint-plugin-sdk/pull/99): Allow use of ${terraform.workspace} in tests

## 0.7.1 (2021-01-10)

### BugFixes

- [#93](https://github.com/terraform-linters/tflint-plugin-sdk/pull/93): tflint: Add workaround for parsing heredoc expressions

## 0.7.0 (2021-01-03)

This release changes the Runner interface. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.23+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/23 for an example of upgrading the SDK. 

### Breaking Changes

- [#83](https://github.com/terraform-linters/tflint-plugin-sdk/pull/83): tflint: Add wantType argument to EvaluateExpr
- [#92](https://github.com/terraform-linters/tflint-plugin-sdk/pull/92): Bump protocol version

### Enhancements

- [#79](https://github.com/terraform-linters/tflint-plugin-sdk/pull/79): tflint: Extend runner API for accessing the root provider configuration
- [#82](https://github.com/terraform-linters/tflint-plugin-sdk/pull/82): tflint: Add support for fetching rule config
- [#85](https://github.com/terraform-linters/tflint-plugin-sdk/pull/85): tflint: Add `IsNullExpr` API
- [#88](https://github.com/terraform-linters/tflint-plugin-sdk/pull/88): Avoid to read file directly from plugin side
- [#91](https://github.com/terraform-linters/tflint-plugin-sdk/pull/91): Make terraform configuration compatible with v0.14

### Chores

- [#80](https://github.com/terraform-linters/tflint-plugin-sdk/pull/80): 
Bump github.com/google/go-cmp from 0.5.3 to 0.5.4
- [#86](https://github.com/terraform-linters/tflint-plugin-sdk/pull/86): Bump github.com/zclconf/go-cty from 1.7.0 to 1.7.1
- [#87](https://github.com/terraform-linters/tflint-plugin-sdk/pull/87): Bump github.com/hashicorp/hcl/v2 from 2.7.1 to 2.8.1
- [#90](https://github.com/terraform-linters/tflint-plugin-sdk/pull/90): Revise README

## 0.6.0 (2020-11-23)

This release adds support for JSON configuration syntax. This is a breaking change and all plugins need to be built using this version in order to work with TFLint v0.21+.

See also https://github.com/terraform-linters/tflint-ruleset-template/pull/19 for an example of upgrading the SDK. 

### Breaking Changes

- [#72](https://github.com/terraform-linters/tflint-plugin-sdk/pull/72): Allow serving custom RuleSet by plugins
  - Change `tflint.Ruleset` struct to the interface. The previous behavior can be reproduced by using the `tflint.BuiltinRuleSet`. If you do not need plugin-specific processing, please use `tflint.BuiltinRuleSet` directly.
- [#78](https://github.com/terraform-linters/tflint-plugin-sdk/pull/78): Bump protocol version

### Enhancements

- [#69](https://github.com/terraform-linters/tflint-plugin-sdk/pull/69): Add support for JSON configuration syntax
- [#75](https://github.com/terraform-linters/tflint-plugin-sdk/pull/75): helper: Update helper runner 

### Chores

- [#68](https://github.com/terraform-linters/tflint-plugin-sdk/pull/68): Bump actions/setup-go from v2.1.2 to v2.1.3
- [#71](https://github.com/terraform-linters/tflint-plugin-sdk/pull/71): Bump github.com/zclconf/go-cty from 1.6.1 to 1.7.0
- [#73](https://github.com/terraform-linters/tflint-plugin-sdk/pull/73): Bump github.com/hashicorp/go-hclog from 0.14.1 to 0.15.0
- [#74](https://github.com/terraform-linters/tflint-plugin-sdk/pull/74): Bump github.com/hashicorp/go-plugin from 1.3.0 to 1.4.0
- [#76](https://github.com/terraform-linters/tflint-plugin-sdk/pull/76): Bump github.com/google/go-cmp from 0.5.2 to 0.5.3
- [#77](https://github.com/terraform-linters/tflint-plugin-sdk/pull/77): Bump github.com/hashicorp/hcl/v2 from 2.7.0 to 2.7.1

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
