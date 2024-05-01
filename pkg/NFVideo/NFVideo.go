package NFVideo

import (
	"errors"
	"fyne.io/fyne/v2"
	"go.novellaforge.dev/novellaforge/data/assets"
	"go.novellaforge.dev/novellaforge/pkg/NFAudio"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var AudioTrack *NFAudio.SpeakerTrack

func init() {
	var err error
	AudioTrack, err = NFAudio.NewSpeakerTrack("video")
	if err != nil {
		return
	}
}

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

// FormatVideo formats the video into a gif and mp3
//
// TODO: Set up an editor integration to allow doing this automatically a build time(No loops)
// or manually at any time allowing the user to set the loop count
func FormatVideo(path string, maxFPS float64, loopCount int, scale scale) error {
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

	//Probe the video
	probe, err := ProbeVideo(absPath)
	if err != nil {
		log.Println("Could not probe video: ", err)
		return err
	}

	totalFrames, err := strconv.Atoi(probe.Streams[0].NbFrames)
	if err != nil {
		log.Println("Could not convert total frames: ", err)
		return err
	}
	log.Println("Total frames: ", totalFrames)

	//Get the frame rate from the real seconds and real frames
	rFrameRate := probe.Streams[0].RFrameRate
	rFrameSplit := strings.Split(rFrameRate, "/")
	rFrameNum := rFrameSplit[0]
	realFrames, err := strconv.Atoi(rFrameNum)
	if err != nil {
		log.Println("Could not convert real frames: ", err)
		return err
	}
	rFrameSeconds := rFrameSplit[1]
	realSeconds, err := strconv.Atoi(rFrameSeconds)
	if err != nil {
		log.Println("Could not convert real seconds: ", err)
		return err
	}

	//Calculate the target fps
	targetFPS := min(float64(realFrames)/float64(realSeconds), maxFPS)

	//Check if a file with the same name but .gif extension already exists if it does loop adding a number to the end of the file name
	gifPath := strings.TrimSuffix(absPath, filepath.Ext(absPath)) + ".gif"
	for i := 1; ; i++ {
		_, err := os.Stat(gifPath)
		if os.IsNotExist(err) {
			break
		}
		gifPath = strings.TrimSuffix(absPath, filepath.Ext(absPath)) + "_" + strconv.Itoa(i) + ".gif"
	}

	/*
		err = CreateGifFromExtractedFrames(absPath, targetFPS, loopCount, gifPath)
		if err != nil {
			log.Println("Could not create gif from video: ", err)
			return err
		}
	*/

	newGif, err := CreateGifFromVideo(absPath, targetFPS, scale)
	if err != nil {
		log.Println("Could not create gif from video: ", err)
		return err
	}

	log.Println("Delay: ", newGif.Delay)

	err = SaveGifWithLoopCount(newGif, loopCount, gifPath)
	if err != nil {
		log.Println("Could not save gif: ", err)
		return err
	}

	mp3Path := strings.TrimSuffix(absPath, filepath.Ext(absPath)) + ".mp3"
	for i := 1; ; i++ {
		_, err := os.Stat(mp3Path)
		if os.IsNotExist(err) {
			break
		}
		mp3Path = strings.TrimSuffix(absPath, filepath.Ext(absPath)) + "_" + strconv.Itoa(i) + ".mp3"
	}
	err = ExtractAudioToMP3(absPath, mp3Path)
	if err != nil {
		log.Println("Could not extract audio: ", err)
		return err
	}

	return nil
}
