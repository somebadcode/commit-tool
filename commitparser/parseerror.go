package commitparser

import (
	"fmt"
)

type ParseError struct {
	err error
	Pos int
}

func (err ParseError) Error() string {
	return fmt.Sprintf("unexpected character at %d: %s", err.Pos, err.err)
}

func (err ParseError) Unwrap() error {
	return err.err
}
