package NFVideo

import (
	"embed"
	"errors"
	"fyne.io/fyne/v2"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

// TODO These embedded Binaries need to be copied to the game folder instead of being unpacked by the editor,
//
//	since it will be the game unpacking them for normal runtime
//
//go:embed binaries
var embeddedBinaries embed.FS
var ffmpegPath string
var ffprobePath string

func UnpackBinaries(location string) error {
	//Check for the binaries folder inside the location
	binaryFolder := filepath.Join(location, "binaries")

	_, err := os.Stat(binaryFolder)
	if os.IsNotExist(err) {
		err := os.MkdirAll(binaryFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	//Check if the binaries folder is empty if it is not return an error
	//If it is empty, unpack the binaries from the embedded binaries folder

	binaryDir, err := os.ReadDir(binaryFolder)
	if err != nil {
		return err
	}

	if len(binaryDir) == 0 {
		switch runtime.GOOS {
		case "windows":
			err := fs.WalkDir(embeddedBinaries, "binaries/win", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				data, err := embeddedBinaries.ReadFile(path)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(binaryFolder, filepath.Base(path)), data, os.ModePerm)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
		default:
			return errors.New("unsupported OS")
		}
	}
	return nil
}

// CheckBinaries checks if the binaries are present in the location
func CheckBinaries() error {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	defaultLocation := filepath.Join(userConfig, "NovellaForge", "ffmpeg")
	binaryLocation := fyne.CurrentApp().Preferences().StringWithFallback("ffmpegLocation", defaultLocation)

	//Check if the binaries are present
	err = UnpackBinaries(binaryLocation)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		ffmpegPath = filepath.Join(binaryLocation, "binaries", "ffmpeg.exe")
		ffprobePath = filepath.Join(binaryLocation, "binaries", "ffprobe.exe")
	default:
		log.Println("Unsupported OS")
		return errors.New("unsupported OS")
	}

	//Check if the ffmpeg and ffprobe binaries are present
	_, err = os.Stat(ffmpegPath)
	if err != nil {
		return err
	}
	_, err = os.Stat(ffprobePath)
	if err != nil {
		return err
	}

	//Check ffmpeg version
	cmd := exec.Command(ffmpegPath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	//Check ffprobe version
	cmd = exec.Command(ffprobePath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

// ParseVideo parses a video file and returns the metadata
func ParseVideo(file string) (*Video, error) {
	//Check if the file exists
	_, err := os.Stat(file)
	if err != nil {
		return &Video{}, nil
	}

	err = CheckBinaries()
	if err != nil {
		return &Video{}, nil
	}

	absLocation, err := filepath.Abs(file)
	if err != nil {
		return &Video{}, nil
	}
	probeArgs := []string{"-show_format", "-show_streams", "-print_format", "json", "-v", "quiet"}

	probeInfo, err := ProbeMP4(absLocation, probeArgs...)
	if err != nil {
		return &Video{}, nil
	}

	i, err := strconv.Atoi(probeInfo.Streams[0].NbFrames)
	if err != nil {
		log.Println("Could not convert the number of frames to an integer")
		return &Video{}, nil
	}

	log.Println("Duration: ", probeInfo.Format.Duration)
	log.Println("Bitrate: ", probeInfo.Format.BitRate)
	log.Println("Number of frames: ", i)
	log.Println("Real Frame Rate: ", probeInfo.Streams[0].RFrameRate)

	video, err := NewVideo(probeInfo, 10)
	if err != nil {
		log.Println("Could not create a new video object")
		return &Video{}, nil
	}

	return video, nil
}
