package addrs

// BareRef is the address of a pseudo bare reference like `ignore_changes = [tags]`.
// This is not a valid address in Terraform, but in the context of TFLint static analysis,
// it may be preferable that any bare reference be parsable.
type BareRef struct {
	referenceable
	Name string
}

func (br BareRef) String() string {
	return br.Name
}
