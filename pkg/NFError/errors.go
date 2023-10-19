package NFError

import "errors"

// All the NFerror types defined here are custom to NovellaForge.
var (
	//ErrMissingArgument is returned when a required argument is missing.
	ErrMissingArgument = errors.New("missing argument")
	//ErrInvalidArgument is returned when an argument is invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	//ErrNotImplemented is returned when a method is not implemented.
	ErrNotImplemented = errors.New("not implemented")
	//ErrProjectNotFound is returned when a project is not found.
	ErrProjectNotFound = errors.New("project not found")
	//ErrNoProjects is returned when there are no projects.
	ErrNoProjects = errors.New("no projects found")
	//ErrProjectAlreadyExists is returned when a project already exists.
	ErrProjectAlreadyExists = errors.New("project already exists")
)
