package NFError

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
)

var (
	ErrConfigLoad              = errors.New("error loading config")
	ErrConfigSave              = errors.New("error saving config")
	ErrFileGet                 = errors.New("error getting file")
	ErrNoProjects              = errors.New("no projects found")
	ErrInvalidArgument         = errors.New("invalid argument")
	ErrKeyAlreadyExists        = errors.New("key already exists")
	ErrMissingArgument         = errors.New("missing argument")
	ErrTypeMismatch            = errors.New("type mismatch")
	ErrKeyNotFound             = errors.New("key not found")
	ErrNotImplemented          = errors.New("not implemented")
	ErrProjectNotFound         = errors.New("project not found")
	ErrProjectAlreadyExists    = errors.New("project already exists")
	ErrSceneSave               = errors.New("error saving scene")
	ErrSceneValidation         = errors.New("scene validation failed")
	ErrCriticalSceneValidation = errors.New("critical scene validation failure")
	ErrNotFound                = errors.New("not found")
	ErrWidgetParse             = errors.New("error parsing widget")
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

func NewErrFileGet(file, reason string) error {
	return fmt.Errorf("%w: %s: %s", ErrFileGet, file, reason)
}

func NewErrConfigLoad(reason string) error {
	return fmt.Errorf("%w: %s", ErrConfigLoad, reason)
}

func NewErrConfigSave(s string) error {
	return fmt.Errorf("%w: %s", ErrConfigSave, s)
}

func NewErrSceneValidation(s string) error {
	return fmt.Errorf("%w: %s", ErrSceneValidation, s)
}

func NewErrCriticalSceneValidation(s string) error {
	return fmt.Errorf("%w: %s", ErrCriticalSceneValidation, s)
}

func NewErrSceneSave(s string) error {
	return fmt.Errorf("%w: %s", ErrSceneSave, s)
}

func NewErrNotFound(s string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, s)
}

func NewErrWidgetParse(widgetName, widgetType string, widgetUUID uuid.UUID, reason string) error {
	return fmt.Errorf("%w: %s of type %s with UUID %v: %s", ErrWidgetParse, widgetName, widgetType, widgetUUID, reason)
}
