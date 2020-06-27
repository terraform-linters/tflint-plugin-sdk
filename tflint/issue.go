package tflint

// List of issue severity levels. The rules implemented by a plugin can be set to any severity.
const (
	// ERROR is possible errors
	ERROR = "Error"
	// WARNING doesn't cause problem immediately, but not good
	WARNING = "Warning"
	// NOTICE is not important, it's mentioned
	NOTICE = "Notice"
)
