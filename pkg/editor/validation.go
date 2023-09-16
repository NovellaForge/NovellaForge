package editor

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/pkg/browser"
	"log"
	"os"
	"os/exec"
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

	// All checks passed, return true and nil error
	return true, nil
}

func CheckAndInstallDependencies(window fyne.Window, terminalWindow fyne.Window) {
	//Check if Go is installed using exec.Command("go", "version")
	//If it is not installed, prompt the user to install it
	goCmd := exec.Command("go", "version")
	err := goCmd.Run()

	if err != nil {
		log.Printf("Go is not installed, prompting user to install it")
		dialog.ShowConfirm("Go is not installed", "Go is not installed on your system. Would you like to install it now? \nGo is required for NovellaForge, if you choose not to install it the editor will close ", func(b bool) {
			if !b {
				os.Exit(0)
			} else {
				//Open the Go download page in the default browser
				err = OpenBrowser("https://go.dev/dl/")
				if err != nil {
					newDialog := dialog.NewError(err, window)
					newDialog.SetOnClosed(func() {
						os.Exit(0)
					})
					newDialog.Show()
					return
				}

				newInfo := dialog.NewInformation("Redirected To Install", "You should have been redirected to GO's installation page, if not please go here https://go.dev/doc/install and install go before before reopening NovellaForge", window)
				newInfo.SetOnClosed(func() {
					os.Exit(0)
				})
				newInfo.Show()
				return

				//TODO: Add in an automatic install of Go

			}

		}, window)
	}

}

func OpenBrowser(s string) error {
	//Open the url in the default browser
	err := browser.OpenURL(s)
	if err != nil {
		return err
	}
	return nil
}
