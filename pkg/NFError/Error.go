package NFError

import (
	"errors"
	"fmt"
)

var (
	ErrNoProjects           = errors.New("no projects found")
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrKeyAlreadyExists     = errors.New("key already exists")
	ErrMissingArgument      = errors.New("missing argument")
	ErrTypeMismatch         = errors.New("type mismatch")
	ErrKeyNotFound          = errors.New("key not found")
	ErrNotImplemented       = errors.New("not implemented")
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
)

func NewErrInvalidArgument(arg, reason string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidArgument, arg, reason)
}

func NewErrKeyAlreadyExists(key string) error {
	return fmt.Errorf("%w: %s", ErrKeyAlreadyExists, key)
}

func NewErrMissingArgument(interfaceName string, key string) error {
	return fmt.Errorf("%w: interface %s missing argument: %s", ErrMissingArgument, interfaceName, key)
}

func NewErrTypeMismatch(expected, actual string) error {
	return fmt.Errorf("%w: type %s does not match type %s", ErrTypeMismatch, expected, actual)
}

func NewErrKeyNotFound(key string) error {
	return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
}

func NewErrNotImplemented(method string) error {
	return fmt.Errorf("%w: %s", ErrNotImplemented, method)
}

func NewErrProjectNotFound(project string) error {
	return fmt.Errorf("%w: %s", ErrProjectNotFound, project)
}

func NewErrProjectAlreadyExists(project string) error {
	return fmt.Errorf("%w: %s", ErrProjectAlreadyExists, project)
}
