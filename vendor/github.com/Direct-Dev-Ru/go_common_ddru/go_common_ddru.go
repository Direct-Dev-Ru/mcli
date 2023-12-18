package go_common_ddru

var Version string = "1.0.1"

type MyType struct {
	Value int
}

type ContextKey string

// NewMyType creates a new instance of MyType.
func NewMyType(value int) MyType {
	return MyType{Value: value}
}

// GetValue returns the value of MyType.
func (mt MyType) GetValue() int {
	return mt.Value
}
