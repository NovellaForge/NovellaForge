package NFVideo

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/data/assets"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

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
			err := fs.WalkDir(assets.BinaryFS, "binaries/win", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				data, err := assets.BinaryFS.ReadFile(path)
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
		FfmpegPath = filepath.Join(binaryLocation, "binaries", "ffmpeg.exe")
		ProbePath = filepath.Join(binaryLocation, "binaries", "ffprobe.exe")
	default:
		log.Println("Unsupported OS")
		return errors.New("unsupported OS")
	}

	//Check if the ffmpeg and ffprobe binaries are present
	_, err = os.Stat(FfmpegPath)
	if err != nil {
		return errors.New("ffmpeg binary not found in the specified location")
	}
	_, err = os.Stat(ProbePath)
	if err != nil {
		return errors.New("ffprobe binary not found in the specified location")
	}

	//Check ffmpeg version
	cmd := exec.Command(FfmpegPath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New("ffmpeg binary is invalid")
	}

	//Check ffprobe version
	cmd = exec.Command(ProbePath, "-version")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New("ffprobe binary is invalid")
	}
	return nil
}

func ParseVideo(path string, maxFPS float64) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Println("Could not get absolute path: ", err)
		return err
	}

	err = CheckBinaries()
	if err != nil {
		log.Println("Could not check binaries: ", err)
		return err
	}

	//Get the number of frames
	probe, err := ProbeVideo(absPath)
	if err != nil {
		log.Println("Could not probe video: ", err)
		return err
	}

	nfvideo := &NFVideo{}
	nbFrames, err := strconv.Atoi(probe.Streams[0].NbFrames)
	if err != nil {
		log.Println("Could not convert number of frames: ", err)
		return err
	}
	nfvideo.TotalFrames = nbFrames

	rFrameRate := probe.Streams[0].RFrameRate
	rFrameSplit := strings.Split(rFrameRate, "/")
	rFrameNum := rFrameSplit[0]
	realFrames, err := strconv.Atoi(rFrameNum)
	if err != nil {
		log.Println("Could not convert real frames: ", err)
		return err
	}
	nfvideo.RealFrames = realFrames
	rFrameSeconds := rFrameSplit[1]
	realSeconds, err := strconv.Atoi(rFrameSeconds)
	if err != nil {
		log.Println("Could not convert real seconds: ", err)
		return err
	}
	nfvideo.RealSeconds = realSeconds

	targetFPS := min(float64(realFrames)/float64(realSeconds), maxFPS)
	nfvideo.TargetFPS = targetFPS

	//Create the folder for the frames
	noExtPath := strings.TrimSuffix(absPath, filepath.Ext(absPath))
	fileName := filepath.Base(noExtPath)

	frameFolder := filepath.Join(noExtPath, fileName+"_frames")
	err = os.MkdirAll(frameFolder, os.ModePerm)
	if err != nil {
		log.Println("Could not create frame folder: ", err)
		return err
	}
	//Write the NFVideo file
	nfVideoPath := filepath.Join(noExtPath, fileName+".NFVideo")
	nfVideoFile, err := os.Create(nfVideoPath)
	defer nfVideoFile.Close()
	if err != nil {
		log.Println("Could not create NFVideo file: ", err)
		return err
	}
	nfVideoData, err := json.Marshal(nfvideo)
	if err != nil {
		log.Println("Could not marshal NFVideo data: ", err)
		return err
	}
	_, err = nfVideoFile.Write(nfVideoData)
	if err != nil {
		log.Println("Could not write NFVideo data: ", err)
		return err
	}
	//Extract the frames
	err = ExtractFrames(absPath, frameFolder, targetFPS)
	if err != nil {
		log.Println("Could not extract frames: ", err)
		return err
	}
	return nil
}
