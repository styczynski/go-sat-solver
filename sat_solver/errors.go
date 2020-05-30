package sat_solver

import "fmt"

type UnsatReason interface {
	Describe() string
}

type UnsatError struct {
	reason UnsatReason
	traceMessage string
}

func (err *UnsatError) Error() string {
	if len(err.traceMessage) > 0 {
		return fmt.Sprintf("%s: The formula cannot be satisified: %s", err.traceMessage, err.reason.Describe())
	}
	return fmt.Sprintf("The formula cannot be satisified: %s", err.reason.Describe())
}

func NewUnsatError(reason UnsatReason) error {
	return &UnsatError{
		reason: reason,
		traceMessage: "",
	}
}

func WrapError(err error, fmtString string, vars... interface{}) error {
	if v, ok := err.(*UnsatError); ok {
		if len(v.traceMessage) > 0 {
			return &UnsatError{
				reason:       v.reason,
				traceMessage: fmt.Sprintf(fmtString+": %s", append(vars, v.traceMessage)...),
			}
		} else {
			return &UnsatError{
				reason:       v.reason,
				traceMessage: fmt.Sprintf(fmtString, vars...),
			}
		}
	}
	return err
}