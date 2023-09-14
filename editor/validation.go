package editor

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

// sanitizeProjectName sanitizes the project name to ensure it's a valid Go identifier
func sanitizeProjectName(name string) (string, error) {
	// Initialize Regular Expression (rgx)
	rgx := regexp.MustCompile(`[^\w\s]`)
	sanitizedName := rgx.ReplaceAllString(name, " ")

	// Remove space
	rgx = regexp.MustCompile(`\s`)
	sanitizedName = rgx.ReplaceAllString(sanitizedName, "")

	// Remove leading non-alphabetic characters
	rgx = regexp.MustCompile(`^[^a-zA-Z]`)
	sanitizedName = rgx.ReplaceAllString(sanitizedName, "")

	// Remove Go-specific keywords and built-in function names if they are at the start of the name
	goKeywords := []string{"break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct", "switch", "type", "var"}
	goBuiltInFuncs := []string{"append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover"}
	for _, keyword := range append(goKeywords, goBuiltInFuncs...) {
		if strings.HasPrefix(sanitizedName, keyword) {
			sanitizedName = strings.Replace(sanitizedName, keyword, "", 1)
		}
	}

	// Check for reserved file names (Windows)
	// TODO: may want to expand this list for other operating systems
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(sanitizedName, reserved) {
			return "", errors.New("project name cannot be a reserved name")
		}
	}

	// Check for length limitations
	// TODO: Update the max length according to os specific limitations
	maxLength := 255
	if len(sanitizedName) > maxLength {
		return "", errors.New("project name cannot exceed maximum length")
	}

	if sanitizedName == "" {
		return "", errors.New("project name cannot be empty")
	}

	//lowercase the whole name
	sanitizedName = strings.ToLower(sanitizedName)

	// Return the sanitized name
	return sanitizedName, nil
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
