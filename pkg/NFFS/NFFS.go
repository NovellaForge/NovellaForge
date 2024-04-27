// Package NFFS allows for grouping embedded filesystems into one multiFS,
// and provides a simple interface for reading from them
// and protections for reading from the local game filesystem
package NFFS

import (
	"embed"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

//TODO add in some more security features to prevent directory traversal attacks and other security vulnerabilities
// This package should be an easy to use viable alternative to the standard os and fs packages, while providing
// some ease of mind for end user security

var localFS fs.FS
var localIsValid bool

func init() {
	ex, err := os.Executable()
	if err != nil {
		localIsValid = false
		return
	}
	log.Println("Executable: ", ex)
	//Check if there is a game directory at the root of the executable
	dataDir := filepath.Join(filepath.Dir(ex), "data")
	info, err := os.Stat(dataDir)
	if err != nil {
		localIsValid = false
		return
	}
	log.Println("Data Directory: ", dataDir)
	//Check if it is a directory
	if !info.IsDir() {
		localIsValid = false
		return
	}

	localFS = os.DirFS(dataDir)
	localIsValid = true
}

// Configuration is a configuration struct for the multiFS that allows specifying the filesystems to use
// New File systems can be added by calling EmbedFS on an embed.FS or a fs.FS
type Configuration struct {
	FSNames   []string
	LocalFS   bool
	OnlyLocal bool
}

// NewConfiguration creates a new configuration for the multiFS if local is true it is checked first
// By leaving fsNames empty all filesystems will be checked in no particular order(it loops through a map[string]fs.fs),
// with most functions except Walk
// returning the first matching file, Walk will walk all filesystems if it is called with an empty fsNames
// Or only the specified filesystems if fsNames is not empty
// You can also manually set OnlyLocal to true after creating the configuration to only check the local filesystem
func NewConfiguration(useLocal bool, fsNames ...string) Configuration {
	return Configuration{
		LocalFS: useLocal,
		FSNames: fsNames,
	}
}

// multiFS is a list of filesystems
type multiFS map[string]fs.FS

// Open opens the first matching file in the filesystems for reading
func (m multiFS) Open(path string, config Configuration) (fs.File, error) {
	if (config.LocalFS || config.OnlyLocal) && localIsValid {
		file, err := localFS.Open(path)
		if err == nil {
			return file, nil
		}
	} else if !localIsValid && config.OnlyLocal {
		return nil, errors.New("local filesystem is not valid")
	}
	if !config.OnlyLocal {
		if len(config.FSNames) > 0 {
			for _, name := range config.FSNames {
				//Check if the name exists in the filesystems
				fsys, ok := m[name]
				if !ok {
					continue
				}
				file, err := fsys.Open(path)
				if err == nil {
					return file, nil
				}
			}
		} else {
			for _, fsys := range m {
				file, err := fsys.Open(path)
				if err == nil {
					return file, nil
				}
			}
		}
	}
	return nil, os.ErrNotExist
}

// Stat returns the fileInfo for the first matching file info in the filesystems
func (m multiFS) Stat(path string, config Configuration) (fs.FileInfo, error) {
	if (config.LocalFS || config.OnlyLocal) && localIsValid {
		fileInfo, err := fs.Stat(localFS, path)
		if err == nil {
			return fileInfo, nil
		}
	} else if !localIsValid && config.OnlyLocal {
		return nil, errors.New("local filesystem is not valid")
	}
	if !config.OnlyLocal {
		if len(config.FSNames) > 0 {
			for _, name := range config.FSNames {
				//Check if the name exists in the filesystems
				fsys, ok := m[name]
				if !ok {
					continue
				}
				file, err := fsys.Open(path)
				if err == nil {
					fileInfo, err := file.Stat()
					if err == nil {
						return fileInfo, nil
					}
				}
			}
		} else {
			for _, fsys := range m {
				file, err := fsys.Open(path)
				if err == nil {
					fileInfo, err := file.Stat()
					if err == nil {
						return fileInfo, nil
					}
				}
			}
		}
	}
	return nil, os.ErrNotExist
}

// ReadFile reads the first matching file in the filesystems and returns the contents as a byte slice
func (m multiFS) ReadFile(path string, config Configuration) ([]byte, error) {
	if (config.LocalFS || config.OnlyLocal) && localIsValid {
		data, err := fs.ReadFile(localFS, path)
		if err == nil {
			return data, nil
		}
	} else if !localIsValid && config.OnlyLocal {
		return nil, errors.New("local filesystem is not valid")
	}
	if !config.OnlyLocal {
		if len(config.FSNames) > 0 {
			for _, name := range config.FSNames {
				//Check if the name exists in the filesystems
				fsys, ok := m[name]
				if !ok {
					continue
				}
				data, err := fs.ReadFile(fsys, path)
				if err == nil {
					return data, nil
				}
			}
		} else {
			for _, fsys := range m {
				data, err := fs.ReadFile(fsys, path)
				if err == nil {
					return data, nil
				}
			}
		}
	}
	return nil, os.ErrNotExist
}

// ReadDir reads the first matching directory in the filesystems and returns a list of directory entries
func (m multiFS) ReadDir(path string, config Configuration) ([]fs.DirEntry, error) {
	if (config.LocalFS || config.OnlyLocal) && localIsValid {
		data, err := fs.ReadDir(localFS, path)
		if err == nil {
			return data, nil
		}
	} else if !localIsValid && config.OnlyLocal {
		return nil, errors.New("local filesystem is not valid")
	}
	if !config.OnlyLocal {
		if len(config.FSNames) > 0 {
			for _, name := range config.FSNames {
				//Check if the name exists in the filesystems
				fsys, ok := m[name]
				if !ok {
					continue
				}
				data, err := fs.ReadDir(fsys, path)
				if err == nil {
					return data, nil
				}
			}
		} else {
			for _, fsys := range m {
				data, err := fs.ReadDir(fsys, path)
				if err == nil {
					return data, nil
				}
			}
		}
	}
	return nil, os.ErrNotExist
}

// Walk walks each filesystem in the multiFS performing the walkFn on each file or directory
func (m multiFS) Walk(dir string, walkFn fs.WalkDirFunc, config Configuration) error {
	var err error
	if (config.LocalFS || config.OnlyLocal) && localIsValid {
		walkErr := fs.WalkDir(localFS, dir, walkFn)
		if walkErr != nil {
			errors.Join(err, walkErr)
		}
	} else if !localIsValid && config.OnlyLocal {
		errors.Join(err, errors.New("local filesystem is not valid"))
	}
	if !config.OnlyLocal {
		if len(config.FSNames) > 0 {
			for _, name := range config.FSNames {
				//Check if the name exists in the filesystems
				fsys, ok := m[name]
				if !ok {
					continue
				}
				walkErr := fs.WalkDir(fsys, dir, walkFn)
				if walkErr != nil {
					errors.Join(err, walkErr)
				}
			}
		} else {
			for _, fsys := range m {
				walkErr := fs.WalkDir(fsys, dir, walkFn)
				if walkErr != nil {
					errors.Join(err, walkErr)
				}
			}
		}
	}
	return err
}

var embeddedFS multiFS

// EmbedFS sets the embedded filesystem to use for loading files
// This function can be called multiple times to add multiple embedded filesystems
// However, please note that unless you use Walk, the functions acting on these added filesystems
// will only return the first matching result. So make sure your filenames/directory names are unique between the filesystems
// Or specify the filesystem to use when calling the function
func EmbedFS(fs embed.FS, names ...string) {
	//Check if a name is provided otherwise set name to "embedded" plus a number until it is unique
	name := "embedded"
	if len(names) > 0 {
		name = names[0]
	}
	for i := 0; embeddedFS[name] != nil; i++ {
		name = "embedded" + strconv.Itoa(i)
	}
	embeddedFS[name] = fs
}

// Walk is an extension of fs.WalkDir for any specified fileSystems
// that have been embedded with EmbedFS
// or all of them if no names are specified(It will return the first matching file)
// and optionally the local game filesystem
// Configuration can be created with NewConfiguration
func Walk(dir string, config Configuration, walkFn fs.WalkDirFunc) error {
	return embeddedFS.Walk(dir, walkFn, config)
}

// ReadFile is an extension of fs.ReadFile for any specified fileSystems
// that have been embedded with EmbedFS
// or all of them if no names are specified(It will return the first matching file)
// and optionally the local game filesystem
// Configuration can be created with NewConfiguration
func ReadFile(path string, config Configuration) ([]byte, error) {
	return embeddedFS.ReadFile(path, config)
}

// ReadDir is an extension of fs.ReadDir for any specified fileSystems
// that have been embedded with EmbedFS
// or all of them if no names are specified(It will return the first matching file)
// and optionally the local game filesystem
// Configuration can be created with NewConfiguration
func ReadDir(path string, config Configuration) ([]fs.DirEntry, error) {
	return embeddedFS.ReadDir(path, config)
}

// Open is an extension of os.Open for any specified fileSystems
// that have been embedded with EmbedFS
// or all of them if no names are specified(It will return the first matching file)
// and optionally the local game filesystem
// Configuration can be created with NewConfiguration
func Open(path string, config Configuration) (fs.File, error) {
	return embeddedFS.Open(path, config)
}

// Stat is an extension of fs.Stat for any specified fileSystems
// that have been embedded with EmbedFS
// or all of them if no names are specified(It will return the first matching file)
// and optionally the local game filesystem
// Configuration can be created with NewConfiguration
func Stat(path string, config Configuration) (fs.FileInfo, error) {
	return embeddedFS.Stat(path, config)
}
