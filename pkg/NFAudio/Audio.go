package NFAudio

import (
	"bytes"
	"io"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"os"
	"time"
)

var done = make(chan bool)
var pauseChannel = make(chan bool)
var resumeChannel = make(chan bool)
var clear = make(chan bool)
var volumeChange = make(chan float64)

var speakerInitialized = false

// Play an audio from bytes with a specified volume, playback speed, and looping option.
// A volume of 0 mutes the audio, and volume increases as the value increases.
// The speed should be a ratio for the speed adjustment, with 1 using the original speed of the file.
// The looping option allows for the audio track to be repeated indefinitely.
func PlayAudioFromBytes(data []byte, volume float64, speed float64, looping bool) error {
	reader := bytes.NewReader(data)
	rc := io.NopCloser(reader)

	return playAudio(rc, volume, speed, looping)
}

// Play an audio file with a specified volume, playback speed, and looping option.
// A volume of 0 mutes the audio, and volume increases as the value increases.
// The speed should be a ratio for the speed adjustment, with 1 using the original speed of the file.
// The looping option allows for the audio track to be repeated indefinitely.
func PlayAudioFromFile(file string, volume float64, speed float64, looping bool) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	return playAudio(f, volume, speed, looping)
}

// Play an audio file with a specified volume, playback speed, and looping option.
// A volume of 0 mutes the audio, and volume increases as the value increases.
// The speed should be a ratio for the speed adjustment, with 1 using the original speed of the file.
// The looping option allows for the audio track to be repeated indefinitely.
func playAudio(rc io.ReadCloser, volume float64, speed float64, looping bool) error {

	streamer, format, err := mp3.Decode(rc)
	if err != nil {
		return err
	}
	defer streamer.Close()

	if !speakerInitialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		speakerInitialized = true
	}
	// Allow for looping of the audio track
	loop := 1
	if looping {
		loop = -1
	}

	loopStreamer := beep.Loop(loop, streamer)

	// Resample the audio track to 44100 Hz
	resampled := beep.Resample(4, format.SampleRate, 44100, loopStreamer)

	// Adjust the audio for the volume.
	// The volume value should be a ratio for the volume adjustment.
	// if volume is 0, then mute the audio
	volumeStreamer := &effects.Volume{
		Streamer: resampled,
		Base:     2,
		Volume:   volume - 5,
		Silent:   volume == 0,
	}

	// Allow for the speed of the track to be changed.
	// The speed value should be a ratio for the speed adjustment. A value of 1 is the original speed.
	speedStreamer := beep.ResampleRatio(4, speed, volumeStreamer)

	// Allow for playing and pausing the audio track
	ctrl := &beep.Ctrl{Streamer: speedStreamer, Paused: false}

	// Create a channel that returns when the player is done playing the audio track.

	// Play the audio track.
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		done <- true
	})))

	for {
		select {
		case <-done:
			return nil
		case <-pauseChannel:
			ctrl.Paused = true
		case <-resumeChannel:
			ctrl.Paused = false
		case <-clear:
			speaker.Clear()
			return nil
		case <-time.After(time.Second):
			// speaker.Lock()
			// fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
			// speaker.Unlock()
		case v := <-volumeChange:
			volumeStreamer.Volume = v - 5
		}
	}

}

// Stop the audio track from playing.
func StopAudio() {
	clear <- true
}

// Pause the audio track.
func PauseAudio() {
	pauseChannel <- true
}

// Resume the audio track.
func ResumeAudio() {
	resumeChannel <- true
}

// Change the volume of the audio track.
func ChangeVolume(volume float64) {
	volumeChange <- volume
}
