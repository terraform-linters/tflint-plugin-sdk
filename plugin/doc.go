// Package plugin contains the implementations needed to make
// the built binary act as a plugin.
//
// A plugin is implemented as an RPC server and the host acts
// as the client, sending analysis requests to the plugin.
// Note that the server-client relationship here is the opposite of
// the communication that takes place during the checking phase.
//
// Implementation details are hidden in go-plugin. This package is
// essentially a wrapper for go-plugin.
package plugin
