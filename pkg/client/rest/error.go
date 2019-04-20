package rest

import "fmt"

type ResponeError struct {
	URL        string
	Method     string
	StatusCode int
}

func (e *ResponeError) Error() string {
	return fmt.Sprintf("%v %v: %d", e.Method, e.URL, e.StatusCode)
}
