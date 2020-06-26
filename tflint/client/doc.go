// Package client contains the implementations required for plugins
// to act as a client. Developers typically query via the Runner interface,
// so they don't need to be aware of this layer.
//
// The client is the actual entity that satisfies the interface.
// It sends a request to the host via RPC, decodes the response,
// and provides APIs for plugins.
//
// Complex structures such as hcl.Expression and hcl.Body are sent/received
// as byte slices and range. Plugins and host parse the byte slice to get the
// original object.
package client
