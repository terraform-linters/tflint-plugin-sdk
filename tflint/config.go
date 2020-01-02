package tflint

// Config is a TFLint configuration applied to a plugin
// At this time, it is not expected that each plugin will reference this directly
type Config struct {
	Rules map[string]*RuleConfig
}

// RuleConfig is a TFLint's rule config
type RuleConfig struct {
	Name    string
	Enabled bool
}
