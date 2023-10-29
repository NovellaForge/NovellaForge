package NFLog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"os"
	"strings"
)

var Directory = ""
var LogDialog *dialog.CustomDialog
var table *widget.Table

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

// CreateLogDialog creates the log dialog
func CreateLogDialog(window fyne.Window) {
	//Create the log text
	label := widget.NewLabel("Click on a log entry to copy it to the clipboard")
	table = widget.NewTableWithHeaders(
		//Length of the table
		func() (int, int) {
			return len(entries), 6
		},
		//Populate the table
		func() fyne.CanvasObject {
			return widget.NewLabel("Template Text")
		},
		//Set the text of the table
		func(i widget.TableCellID, o fyne.CanvasObject) {
			switch i.Col {
			case 0:
				o.(*widget.Label).SetText(entries[i.Row].Date)
			case 1:
				o.(*widget.Label).SetText(entries[i.Row].Time)
			case 2:
				o.(*widget.Label).SetText(entries[i.Row].File)
			case 3:
				o.(*widget.Label).SetText(entries[i.Row].Line)
			case 4:
				o.(*widget.Label).SetText(entries[i.Row].Message)
			}
		},
	)
	table.OnSelected = func(id widget.TableCellID) {
		//Get the entry from the row
		curEntry := entries[id.Row]
		//Create a string with the date, time, file, line, and message
		str := curEntry.Date + " " + curEntry.Time + " " + curEntry.File + " " + curEntry.Line + " " + curEntry.Message
		//Copy the string to the clipboard
		window.Clipboard().SetContent(str)
	}
	vbox := container.NewVBox(label, table)

	LogDialog = dialog.NewCustom("Log", "Close", vbox, window)
	//Resize the log to be the size of the window - 50 pixels but clamped at 500 pixels square
	width := max(window.Canvas().Size().Width-50, 500)
	height := max(window.Canvas().Size().Height-50, 500)
	LogDialog.Resize(fyne.NewSize(width, height))
}

// ShowDialog shows the log dialog
func ShowDialog(window fyne.Window) {
	//If the log dialog is nil, create it
	if LogDialog == nil {
		CreateLogDialog(window)
	}
	//Show the log dialog
	LogDialog.Show()
}

// GetDirectory returns the directory that the log files are stored in
func GetDirectory() string {
	return Directory
}

// SetDirectory sets the directory that the log files are stored in and creates it if it doesn't exist
func SetDirectory(dir string) {
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

func SetUp() error {
	//Make all directories to the log file
	err := os.MkdirAll(Directory, os.ModePerm)
	if err != nil {
		return err
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
		err = os.Rename(Directory+"log.txt", Directory+fileInfo.ModTime().Format("2006-01-02 15:04:05")+".txt")
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

	// Inside SetUp or another function
	go func() {
		for logEntry := range logChannel {
			e := parseLogEntry(logEntry)
			entries = append(entries, e)
			table.Refresh()
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
	if len(fileAndLine) == 2 {
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
