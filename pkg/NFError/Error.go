package NFError

import "errors"

// All the NFError types defined here are custom to NovellaForge.
var (
	//ErrInvalidArgument is returned when an argument is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	//ErrKeyAlreadyExists is returned when a key already exists.
	ErrKeyAlreadyExists = errors.New("key already exists")
	//ErrKeyNotFound is returned when a key is not found.
	ErrKeyNotFound = errors.New("key not found")
	//ErrTypeMismatch is returned when a type mismatch occurs.
	ErrTypeMismatch = errors.New("type mismatch")
	//ErrNotImplemented is returned when a method is not implemented.
	ErrNotImplemented = errors.New("not implemented")
	//ErrProjectNotFound is returned when a project is not found.
	ErrProjectNotFound = errors.New("project not found")
	//ErrNoProjects is returned when there are no projects.
	ErrNoProjects = errors.New("no projects found")
	//ErrProjectAlreadyExists is returned when a project already exists.
	ErrProjectAlreadyExists = errors.New("project already exists")
)

func ErrMissingArgument(interfaceName string, key ...string) error {
	if len(key) > 0 {
		stringList := ""
		for i := 0; i < len(key); i++ {
			stringList += key[i] + ", "
		}
		return errors.New("interface: " + interfaceName + " missing one or more argument(s): " + stringList)
	}
	return errors.New("interface: " + interfaceName + " missing one or more argument(s)")
}
