package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

// Resource is an intermediate representation of terraform.Resource.
type Resource struct {
	Mode         addrs.ResourceMode
	Name         string
	Type         string
	Config       []byte
	ConfigRange  hcl.Range
	Count        []byte
	CountRange   hcl.Range
	ForEach      []byte
	ForEachRange hcl.Range

	ProviderConfigRef *configs.ProviderConfigRef
	Provider          addrs.Provider

	Managed *ManagedResource

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeResource(resource *Resource) (*configs.Resource, hcl.Diagnostics) {
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

	return &configs.Resource{
		Mode:    resource.Mode,
		Name:    resource.Name,
		Type:    resource.Type,
		Config:  file.Body,
		Count:   count,
		ForEach: forEach,

		ProviderConfigRef: resource.ProviderConfigRef,
		Provider:          resource.Provider,

		Managed: managed,

		DeclRange: resource.DeclRange,
		TypeRange: resource.TypeRange,
	}, nil
}

// ManagedResource is an intermediate representation of terraform.ManagedResource.
type ManagedResource struct {
	Connection   *Connection
	Provisioners []*Provisioner

	CreateBeforeDestroy bool
	PreventDestroy      bool
	IgnoreAllChanges    bool

	CreateBeforeDestroySet bool
	PreventDestroySet      bool
}

func decodeManagedResource(resource *ManagedResource) (*configs.ManagedResource, hcl.Diagnostics) {
	if resource == nil {
		return nil, nil
	}

	connection, diags := decodeConnection(resource.Connection)
	if diags.HasErrors() {
		return nil, diags
	}

	provisioners := make([]*configs.Provisioner, len(resource.Provisioners))
	for i, p := range resource.Provisioners {
		provisioner, diags := decodeProvisioner(p)
		if diags.HasErrors() {
			return nil, diags
		}
		provisioners[i] = provisioner
	}

	return &configs.ManagedResource{
		Connection:   connection,
		Provisioners: provisioners,

		CreateBeforeDestroy: resource.CreateBeforeDestroy,
		PreventDestroy:      resource.PreventDestroy,
		IgnoreAllChanges:    resource.IgnoreAllChanges,

		CreateBeforeDestroySet: resource.CreateBeforeDestroySet,
		PreventDestroySet:      resource.PreventDestroySet,
	}, nil
}
