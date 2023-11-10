package NFLog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var Directory = ""
var LogList *widget.List
var LogDialog *dialog.CustomDialog

type entry struct {
	Date    string
	Time    string
	File    string
	Line    string
	Message string
}

var entries []entry

type writer chan<- string

func (c writer) Write(p []byte) (n int, err error) {
	c <- string(p)
	return len(p), nil
}

// ShowDialog shows the log dialog
func ShowDialog(window fyne.Window) {
	//If the log dialog is nil, create it
	if LogDialog == nil {
		err := SetUp(window)
		if err != nil {
			return
		}
	}
	LogDialog.Show()

}

// GetDirectory returns the directory that the log files are stored in
func GetDirectory() string {
	return Directory
}

// SetDirectory sets the directory that the log files are stored in and creates it if it doesn't exist
func SetDirectory(dir string) {
	//Make sure it ends in a slash and only one slash
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	//Check if the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		//Create the directory
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	Directory = dir
}

func SetUp(window fyne.Window, dir ...string) error {
	if len(dir) > 0 {
		SetDirectory(dir[0])
	} else {
		SetDirectory("logs/")
	}

	//Check if a log.txt file already exists
	fileInfo, err := os.Stat(Directory + "log.txt")
	if err != nil {
		//Create a new log.txt file
		_, err = os.Create(Directory + "log.txt")
		if err != nil {
			return err
		}
	}

	//Copy the new log file to one named with its creation date
	if fileInfo != nil {
		err = os.Rename(Directory+"log.txt", Directory+fileInfo.ModTime().Format("2006-01-02-1504")+".txt")
		if err != nil {
			panic(err)
		}
		//Create a new log.txt file
		_, err = os.Create(Directory + "log.txt")
	}

	//Open the log file for writing
	logFile, err := os.OpenFile(Directory+"log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	logChannel := make(chan string)
	channelWriter := writer(logChannel)
	multi := io.MultiWriter(os.Stdout, logFile, channelWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.SetOutput(multi)

	LogList = widget.NewList(
		func() int {
			return len(entries)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {}),
				widget.NewLabel("Template List Item"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			//create a message with everything but the date and time
			message := entries[i].File + ":" + entries[i].Line + " | " + entries[i].Message
			dateTime := entries[i].Date + "-" + entries[i].Time
			//Set the button to copy the message to the clipboard
			o.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {
				window.Clipboard().SetContent(dateTime + " | " + entries[i].File + ":" + entries[i].Line + " | " + entries[i].Message)
			}
			//Set the label to the message
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(message)
			//Set the button text to the date and time
			o.(*fyne.Container).Objects[0].(*widget.Button).SetText(dateTime)
		},
	)

	LogDialog = dialog.NewCustom("Log", "Close", LogList, window)

	//Listen for a screen size change
	window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			LogDialog.Hide()
		}
	})

	go func() {
		for {
			if LogDialog != nil {
				//Get the size of the screen
				size := window.Canvas().Size()
				//Set the size of the dialog to 80% of the screen
				LogDialog.Resize(fyne.NewSize(size.Width*0.8, size.Height*0.8))
			}
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for logEntry := range logChannel {
			e := parseLogEntry(logEntry)
			entries = append(entries, e)
			if LogDialog != nil {
				LogDialog.Refresh()
			}
		}
	}()

	return nil
}

func parseLogEntry(logEntry string) entry {
	// Split the log entry into its parts based on space
	parts := strings.SplitN(logEntry, " ", 4)

	// Create a new entry
	e := entry{}

	// Set the date and time
	e.Date = parts[0]
	e.Time = parts[1]

	// Further split the "file:line" into file and line
	fileAndLine := strings.Split(parts[2], ":")
	if len(fileAndLine) >= 2 {
		e.File = fileAndLine[0]
		e.Line = fileAndLine[1]
	} else {
		e.File = parts[2]  // If it doesn't split, use it as is
		e.Line = "unknown" // We couldn't determine the line number
	}

	// Set the message
	e.Message = parts[3]

	return e
}
