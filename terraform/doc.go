// Package terraform contains structures for Terraform's alternative
// representations. The reason for providing these is to avoid depending
// on Terraform for this plugin.
//
// The structures provided by this package are minimal used for static analysis.
// Also, this is often not transferred directly from the host process, but through
// an intermediate representation. See the package tflint for details.
package terraform
