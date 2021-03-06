package synch

import "fmt"

type synchInitError struct {
	method string
	errMsg string
}

func (e *synchInitError) Error() string {
	return fmt.Sprintf("[ERROR] synch init: %s in method %s", e.errMsg, e.method)
}

type mappingError struct {
	errMsg string
}

func (e *mappingError) Error() string {
	return fmt.Sprintf("[ERROR] mapping: %s", e.errMsg)
}
