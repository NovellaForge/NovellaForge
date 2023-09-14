package editor

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

// sanitizeProjectName sanitizes the project name to ensure it's a valid Go identifier
func sanitizeProjectName(name string) (bool, error) {
	// Initialize Regular Expression (rgx)
	rgx := regexp.MustCompile(`[^\w\s]`)
	if rgx.MatchString(name) {
		return false, errors.New("project name contains invalid characters")
	}
	sanitizedName := rgx.ReplaceAllString(name, " ")

	// Remove space
	rgx = regexp.MustCompile(`\s`)
	if rgx.MatchString(sanitizedName) {
		return false, errors.New("project name contains spaces")
	}
	sanitizedName = rgx.ReplaceAllString(sanitizedName, "")

	// Remove leading non-alphabetic characters
	rgx = regexp.MustCompile(`^[^a-zA-Z]`)
	if rgx.MatchString(sanitizedName) {
		return false, errors.New("project name starts with a non-alphabetic character")
	}
	sanitizedName = rgx.ReplaceAllString(sanitizedName, "")

	// Check for reserved file names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(sanitizedName, reserved) {
			return false, errors.New("project name cannot be a reserved name")
		}
	}

	// Check for length limitations
	maxLength := 255
	if len(sanitizedName) > maxLength {
		return false, errors.New("project name cannot exceed maximum length")
	}

	if sanitizedName == "" {
		return false, errors.New("project name cannot be empty")
	}

	// All checks passed, return true and nil error
	return true, nil
}

// CreateFileIfNotExist creates a file if it does not exist
func CreateFileIfNotExist(filePath string, fileData string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file
		err := os.WriteFile(filePath, []byte(fileData), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
