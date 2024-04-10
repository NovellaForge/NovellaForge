package NFFS

import (
	"embed"
	"errors"
	"go.novellaforge.dev/novellaforge/pkg/NFError"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MultiFS is a list of filesystems
type MultiFS []fs.FS

// Open opens the first matching file in the filesystems for reading
func (m MultiFS) Open(path string) (fs.File, error) {
	for _, fsys := range m {
		file, err := fsys.Open(path)
		if err == nil {
			return file, nil
		}
	}
	return nil, os.ErrNotExist
}

// ReadFile reads the first matching file in the filesystems and returns the contents as a byte slice
func (m MultiFS) ReadFile(name string) ([]byte, error) {
	for _, fsys := range m {
		data, err := fs.ReadFile(fsys, name)
		if err == nil {
			return data, nil
		}
	}
	return nil, os.ErrNotExist
}

// ReadDir reads the first matching directory in the filesystems and returns a list of directory entries
func (m MultiFS) ReadDir(name string) ([]fs.DirEntry, error) {
	for _, fsys := range m {
		data, err := fs.ReadDir(fsys, name)
		if err == nil {
			return data, nil
		}
	}
	return nil, os.ErrNotExist
}

// Walk walks each filesystem in the MultiFS performing the walkFn on each file or directory
func (m MultiFS) Walk(dir string, walkFn fs.WalkDirFunc) error {
	var err error
	for _, fsys := range m {
		walkErr := fs.WalkDir(fsys, dir, walkFn)
		if walkErr != nil {
			errors.Join(err, walkErr)
		}
	}
	return err
}

var embeddedFS MultiFS

// EmbedFS sets the embedded filesystem to use for loading files
// This function can be called multiple times to add multiple embedded filesystems
// However, please note that unless you use Walk, the functions acting on these added filesystems
// will only return the first matching result. So make sure your filenames/directory names are unique between the filesystems
func EmbedFS(fs embed.FS) {
	embeddedFS = append(embeddedFS, fs)
}

// Walk walks both the embedded and game filesystems
func Walk(dir string, walkFn fs.WalkDirFunc) error {
	dir = strings.TrimSpace(dir)
	if !strings.HasPrefix(dir, "game/") {
		dir = "game/" + dir
	}
	embeddedDir := strings.TrimPrefix(dir, "game/")
	if !fs.ValidPath(dir) {
		return NFError.NewErrFileGet(dir, "invalid path")
	}

	var err error
	embedErr := fs.WalkDir(embeddedFS, embeddedDir, walkFn)
	if embedErr != nil {
		errors.Join(err, embedErr)
	}

	localErr := filepath.WalkDir(dir, walkFn)
	if localErr != nil {
		errors.Join(err, localErr)
	}
	return err
}

// Open uses os.Open for the game fileSystem and
// MultiFS.Open for the embedded filesystem(s) to return a file opened for reading
func Open(path string) (fs.File, error) {
	// Trim any leading or trailing white spaces
	path = strings.TrimSpace(path)
	// Check if the path already starts with "game/"
	if !strings.HasPrefix(path, "game") {
		path = "game/" + path
		path = filepath.Clean(path)
	}

	embeddedPath := strings.TrimPrefix(path, "game")
	if !fs.ValidPath(path) {
		return nil, NFError.NewErrFileGet(path, "invalid path")
	}

	file, err := embeddedFS.Open(embeddedPath)
	if err != nil {
		file, err = os.Open(path)
		if err != nil {
			return nil, NFError.NewErrFileGet(path, err.Error())
		}
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, NFError.NewErrFileGet(path, err.Error())
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return nil, NFError.NewErrFileGet(path, "is a symlink which is not supported")
	}

	if fileInfo.IsDir() {
		return nil, NFError.NewErrFileGet(path, "is a directory")
	}

	return file, nil
}
