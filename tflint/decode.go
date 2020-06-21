package tflint

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
)

// Resource is an intermediate representation of configs.Resource.
type Resource struct {
	Mode         terraform.ResourceMode
	Name         string
	Type         string
	Config       []byte
	ConfigRange  hcl.Range
	Count        []byte
	CountRange   hcl.Range
	ForEach      []byte
	ForEachRange hcl.Range

	ProviderConfigRef *terraform.ProviderConfigRef

	Managed *ManagedResource

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeResource(resource *Resource) (*terraform.Resource, hcl.Diagnostics) {
	file, diags := parseConfig(resource.Config, resource.ConfigRange.Filename, resource.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	var count hcl.Expression
	if resource.Count != nil {
		count, diags = parseExpression(resource.Count, resource.CountRange.Filename, resource.CountRange.Start)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	var forEach hcl.Expression
	if resource.ForEach != nil {
		forEach, diags = parseExpression(resource.ForEach, resource.ForEachRange.Filename, resource.ForEachRange.Start)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	managed, diags := decodeManagedResource(resource.Managed)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Resource{
		Mode:    resource.Mode,
		Name:    resource.Name,
		Type:    resource.Type,
		Config:  file.Body,
		Count:   count,
		ForEach: forEach,

		ProviderConfigRef: resource.ProviderConfigRef,

		Managed: managed,

		DeclRange: resource.DeclRange,
		TypeRange: resource.TypeRange,
	}, nil
}

// ManagedResource is an intermediate representation of configs.ManagedResource.
type ManagedResource struct {
	Connection   *Connection
	Provisioners []*Provisioner

	CreateBeforeDestroy bool
	PreventDestroy      bool
	IgnoreAllChanges    bool

	CreateBeforeDestroySet bool
	PreventDestroySet      bool
}

func decodeManagedResource(resource *ManagedResource) (*terraform.ManagedResource, hcl.Diagnostics) {
	connection, diags := decodeConnection(resource.Connection)
	if diags.HasErrors() {
		return nil, diags
	}

	provisioners := make([]*terraform.Provisioner, len(resource.Provisioners))
	for i, p := range resource.Provisioners {
		provisioner, diags := decodeProvisioner(p)
		if diags.HasErrors() {
			return nil, diags
		}
		provisioners[i] = provisioner
	}

	return &terraform.ManagedResource{
		Connection:   connection,
		Provisioners: provisioners,

		CreateBeforeDestroy: resource.CreateBeforeDestroy,
		PreventDestroy:      resource.PreventDestroy,
		IgnoreAllChanges:    resource.IgnoreAllChanges,

		CreateBeforeDestroySet: resource.CreateBeforeDestroySet,
		PreventDestroySet:      resource.PreventDestroySet,
	}, nil
}

// Connection is an intermediate representation of configs.Connection.
type Connection struct {
	Config      []byte
	ConfigRange hcl.Range

	DeclRange hcl.Range
}

func decodeConnection(connection *Connection) (*terraform.Connection, hcl.Diagnostics) {
	if connection == nil {
		return nil, hcl.Diagnostics{}
	}

	file, diags := parseConfig(connection.Config, connection.ConfigRange.Filename, connection.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Connection{
		Config:    file.Body,
		DeclRange: connection.DeclRange,
	}, nil
}

// Provisioner is an intermediate representation of terraform.Provisioner.
type Provisioner struct {
	Type        string
	Config      []byte
	ConfigRange hcl.Range
	Connection  *Connection
	When        terraform.ProvisionerWhen
	OnFailure   terraform.ProvisionerOnFailure

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeProvisioner(provisioner *Provisioner) (*terraform.Provisioner, hcl.Diagnostics) {
	file, diags := parseConfig(provisioner.Config, provisioner.ConfigRange.Filename, provisioner.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	connection, diags := decodeConnection(provisioner.Connection)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Provisioner{
		Type:       provisioner.Type,
		Config:     file.Body,
		Connection: connection,
		When:       provisioner.When,
		OnFailure:  provisioner.OnFailure,

		DeclRange: provisioner.DeclRange,
		TypeRange: provisioner.TypeRange,
	}, nil
}
