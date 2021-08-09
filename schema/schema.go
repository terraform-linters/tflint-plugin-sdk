package schema

type BodySchema struct {
	Attributes []AttributeSchema
	Blocks     []BlockSchema
}

type AttributeSchema struct {
	Name string
}

type BlockSchema struct {
	Type       string
	LabelNames []string

	Body *BodySchema
}
