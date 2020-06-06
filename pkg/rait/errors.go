package rait

import "fmt"

type ErrDecode struct {
	Type string
	Key  string
	Err  error
}

func (e ErrDecode) Unwrap() error {
	return e.Err
}

func (e ErrDecode) Error() string {
	return fmt.Errorf("failed to decode %s of %s: %w", e.Key, e.Type, e.Err).Error()
}

func NewErrDecode(Type string, Key string, Err error) ErrDecode {
	return ErrDecode{
		Type: Type,
		Key: Key,
		Err: Err,
	}
}
