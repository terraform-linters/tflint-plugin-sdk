// Package helper contains implementations for plugin testing.
//
// You can test the implemented rules using the mock Runner that is not
// an gRPC client. It is similar to TFLint's Runner, but is implemented
// from scratch to avoid Terraform dependencies.
//
// Some implementations of the mock Runner have been simplified. As a result,
// note that some features may not behave exactly as they should.
package helper
