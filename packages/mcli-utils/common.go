package mcliutils

// TiifFunc- returns funcTrue or funcFalse result as interface{}, error tulip accordinly by the ifCase bool param
// args ...interface{} is optional args for true and false functions
func TiifFunc(ifCase bool, funcTrue func(args ...interface{}) (interface{}, error),
	funcFalse func(args ...interface{}) (interface{}, error), args ...interface{}) (interface{}, error) {

	if ifCase {
		return funcTrue(args)
	}
	return funcFalse(args)
}

// Tiif - returns trueValue or falseValue passed as arguments of interface{} type
// accordinly by the ifCase bool param.
// Example: nCount := Tiif(bFlag, 1, math.MaxInt).(int)
func Tiif(ifCase bool, trueValue interface{}, falseValue interface{}) interface{} {
	if ifCase {
		return trueValue
	}
	return falseValue
}
