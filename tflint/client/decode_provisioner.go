package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

// Provisioner is an intermediate representation of terraform.Provisioner.
type Provisioner struct {
	Type        string
	Config      []byte
	ConfigRange hcl.Range
	Connection  *Connection
	When        configs.ProvisionerWhen
	OnFailure   configs.ProvisionerOnFailure

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeProvisioner(provisioner *Provisioner) (*configs.Provisioner, hcl.Diagnostics) {
	file, diags := parseConfig(provisioner.Config, provisioner.ConfigRange.Filename, provisioner.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	connection, diags := decodeConnection(provisioner.Connection)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.Provisioner{
		Type:       provisioner.Type,
		Config:     file.Body,
		Connection: connection,
		When:       provisioner.When,
		OnFailure:  provisioner.OnFailure,

		DeclRange: provisioner.DeclRange,
		TypeRange: provisioner.TypeRange,
	}, nil
}

// Connection is an intermediate representation of terraform.Connection.
type Connection struct {
	Config      []byte
	ConfigRange hcl.Range

	DeclRange hcl.Range
}

func decodeConnection(connection *Connection) (*configs.Connection, hcl.Diagnostics) {
	if connection == nil {
		return nil, hcl.Diagnostics{}
	}

	file, diags := parseConfig(connection.Config, connection.ConfigRange.Filename, connection.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.Connection{
		Config:    file.Body,
		DeclRange: connection.DeclRange,
	}, nil
}
