package presto

type OperatorError struct {
	errormsg string
}

func (e *OperatorError) Error() string {
	return e.errormsg
}
