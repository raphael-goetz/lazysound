package app

import "fmt"

type Kind string

const (
	KindConfig     Kind = "config"
	KindAuth       Kind = "auth"
	KindTokenStore Kind = "token_store"
	KindFetch      Kind = "fetch"
)

type Error struct {
	Kind Kind
	Op   string
	Err  error
}

func (e Error) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s %s: %v", e.Kind, e.Op, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

func (e Error) Unwrap() error { return e.Err }

func wrap(kind Kind, op string, err error) error {
	if err == nil {
		return nil
	}
	return Error{Kind: kind, Op: op, Err: err}
}
