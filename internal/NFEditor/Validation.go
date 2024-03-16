package NFEditor

import (
	"errors"
	"fyne.io/fyne/v2"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// SanitizeProjectName sanitizes the project name to ensure it's a valid Go identifier
func SanitizeProjectName(name string) (bool, error) {
	// Initialize Regular Expression (rgx)
	rgx := regexp.MustCompile(`[^\w\s]`)
	if rgx.MatchString(name) {
		return false, errors.New("project name contains invalid characters")
	}
	sanitizedName := rgx.ReplaceAllString(name, " ")

	// Remove space
	rgx = regexp.MustCompile(`\s`)
	if rgx.MatchString(sanitizedName) {
		return false, errors.New("project name cannot contain spaces")
	}
	sanitizedName = rgx.ReplaceAllString(sanitizedName, "")

	// Check for reserved file names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(sanitizedName, reserved) {
			return false, errors.New("project name cannot be a system reserved name")
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

	return true, nil
}

func CheckAndInstallDependencies(window fyne.Window) {
	//Check if Go is installed using exec.Command("go", "version")
	//If it is not installed, prompt the user to install it
	goCmd := exec.Command("go", "version")
	err := goCmd.Run()
	if err != nil {
		log.Printf("Go is not installed, prompting user to install it")

		//TODO: Add in an automatic install of Go

	}

}
