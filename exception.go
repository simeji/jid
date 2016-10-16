package jig

import (
	"fmt"
)

type QueryParseError struct {
	Message string
}

func NewQueryParseError(s string) error {
	return &QueryParseError{
		Message: s,
	}
}

func (e *QueryParseError) Error() string {
	return fmt.Sprintf("Query parse error: %s", e.Message)
}
