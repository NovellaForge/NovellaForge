package editor

import "errors"

// All the error types defined here are custom to NovellaForge.
var (
	//ErrNotImplemented is returned when a method is not implemented.
	ErrNotImplemented = errors.New("not implemented")
	//ErrProjectNotFound is returned when a project is not found.
	ErrProjectNotFound = errors.New("project not found")
	//ErrNoProjects is returned when there are no projects.
	ErrNoProjects = errors.New("no projects found")
	//ErrProjectAlreadyExists is returned when a project already exists.
	ErrProjectAlreadyExists = errors.New("project already exists")
)
