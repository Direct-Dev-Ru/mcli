package mcliutils

import "strings"

type CommonOption struct {
	optionMap map[string]interface{}
}

func (opt *CommonOption) GetOptionMap(optionName string) (interface{}, bool) {

	optionName = strings.TrimSpace(optionName)
	optionName = strings.ReplaceAll(optionName, "-", "_")
	optionName = strings.ToUpper(optionName)

	if val, ok := opt.optionMap[optionName]; !ok {
		return nil, ok
	} else {
		return val, ok
	}
}

func (opt *CommonOption) SetOptionMap(optionName string, optionValue interface{}) interface{} {
	if opt.optionMap == nil {
		opt.optionMap = make(map[string]interface{})
	}
	optionName = strings.TrimSpace(optionName)
	optionName = strings.ReplaceAll(optionName, "-", "_")
	optionName = strings.ToUpper(optionName)

	opt.optionMap[optionName] = optionValue
	return nil
}

func (opt *CommonOption) GetStringOption(optionName string) (string, bool) {
	if val, ok := opt.optionMap[optionName]; !ok {
		return "", ok
	} else {
		strval, ok := val.(string)
		if !ok {
			return "", ok
		}
		return strval, ok
	}
}

func (opt *CommonOption) GetIntOption(optionName string) (int, bool) {
	if val, ok := opt.optionMap[optionName]; !ok {
		return 0, ok
	} else {
		intval, ok := val.(int)
		if !ok {
			return 0, ok
		}
		return intval, ok
	}
}

func (opt *CommonOption) GetBoolOption(optionName string) (bool, bool) {
	if val, ok := opt.optionMap[optionName]; !ok {
		return false, ok
	} else {
		bval, ok := val.(bool)
		if !ok {
			return false, ok
		}
		return bval, ok
	}
}
