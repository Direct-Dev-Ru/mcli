package mclierror

import "fmt"

type Status int

const (
	Undefined Status = iota
	ResourceNotFound
	InvalidParameter
	InvalidLogin
	NotAuthorized
)

var ResourceNotFoundError CommonError = CommonError{Status: ResourceNotFound}
var InvalidParameterError CommonError = CommonError{Status: InvalidParameter}
var InvalidLoginError CommonError = CommonError{Status: InvalidLogin}

type CommonError struct {
	Status  Status
	Message string
	Err     error
}

func NewError(message string, status ...Status) error {
	var st Status
	if len(status) > 0 {
		st = status[0]
	}
	return CommonError{Status: st, Message: message}
}

func NewCommonError(message string, status ...Status) CommonError {
	var st Status
	if len(status) > 0 {
		st = status[0]
	}
	return CommonError{Status: st, Message: message}
}

func NewCommonErrorFromError(err error) CommonError {
	if err != nil {
		return CommonError{Status: Undefined, Message: err.Error()}
	}
	return CommonError{}
}

func (ce CommonError) Error() string {
	if ce2, ok := ce.Err.(CommonError); ok && ce2.IsNil() {
		return ce.Message
	}
	if ce2, ok := ce.Err.(CommonError); ok {
		return fmt.Sprintf("%s: %s", ce.Message, ce2.Error())
	} else if !ok {
		return fmt.Sprintf("%s: %s", ce.Message, ce.Err.Error())
	}
	return ce.Message
}

func (ce *CommonError) Wrap(err error) {
	ce.Err = err
}

func (ce CommonError) UnWrap() error {
	return ce.Err
}
func (ce CommonError) IsNil() bool {
	return len(ce.Message) == 0 && ce.Status == Undefined
}

func (ce CommonError) Is(target error) bool {
	if ceTarget, ok := target.(CommonError); ok {
		if ce.Status == ceTarget.Status {
			return true
		}
		if ceInner, ok := ce.Err.(CommonError); ok {

			if !ceInner.IsNil() {
				return ceInner.Is(target)
			}
		}
	}
	return false
}
