package NFAudio

import (
	"bytes"
	"errors"
	"io"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"os"
	"time"
)

var speakerInitialized = false

// var masterVolumeChannel = make(chan float64)
// var muteChannel = make(chan bool)
// var unmuteChannel = make(chan bool)
var masterVolume = 5.0
var SpeakerTracks = make(map[string]*SpeakerTrack)

type SpeakerTrack struct {
	name                string
	state               string
	done                chan bool
	pauseChannel        chan bool
	resumeChannel       chan bool
	clear               chan bool
	volumeChange        chan float64
	muteChannel         chan bool
	unmuteChannel       chan bool
	masterVolumeChannel chan float64
}

func NewSpeakerTrack(name string) (*SpeakerTrack, error) {
	if _, ok := SpeakerTracks[name]; !ok {
		SpeakerTracks[name] = &SpeakerTrack{
			name:                name,
			state:               "created",
			done:                make(chan bool),
			pauseChannel:        make(chan bool),
			resumeChannel:       make(chan bool),
			clear:               make(chan bool),
			muteChannel:         make(chan bool),
			unmuteChannel:       make(chan bool),
			volumeChange:        make(chan float64),
			masterVolumeChannel: make(chan float64),
		}
		return SpeakerTracks[name], nil
	} else {
		return nil, errors.New("speakerTrack already exists")
	}
}

// var done = make(chan bool)
// var pauseChannel = make(chan bool)
// var resumeChannel = make(chan bool)
// var clear = make(chan bool)
// var volumeChange = make(chan float64)

// Play an audio from bytes with a specified volume, playback speed, and looping option.
// A volume of 0 mutes the audio, and volume increases as the value increases.
// The speed should be a ratio for the speed adjustment, with 1 using the original speed of the file.
// The loops option allows for the audio track to be repeated indefinitely. Choose -1 for infinitely looping audio
// A loops value of 0 and 1 have the same functionality, playing the audio once.
func (s *SpeakerTrack) PlayAudioFromBytes(data []byte, volume float64, speed float64, loops int) error {
	reader := bytes.NewReader(data)
	rc := io.NopCloser(reader)
	return s.playAudio(rc, volume, speed, loops)
}

// Play an audio file with a specified volume, playback speed, and looping option.
// A volume of 0 mutes the audio, and volume increases as the value increases.
// The speed should be a ratio for the speed adjustment, with 1 using the original speed of the file.
// The loops option allows for the audio track to be repeated indefinitely. Choose -1 for infinitely looping audio
// A loops value of 0 and 1 have the same functionality, playing the audio once.
func (s *SpeakerTrack) PlayAudioFromFile(file string, volume float64, speed float64, loops int) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	return s.playAudio(f, volume, speed, loops)
}

// Play an audio file with a specified volume, playback speed, and looping option.
// A volume of 0 mutes the audio, and volume increases as the value increases.
// The speed should be a ratio for the speed adjustment, with 1 using the original speed of the file.
// The looping option allows for the audio track to be repeated indefinitely. Choose -1 for infinitely looping audio
// A loops value of 0 and 1 have the same functionality, playing the audio once.
func (s *SpeakerTrack) playAudio(rc io.ReadCloser, volume float64, speed float64, loops int) error {

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
	numLoops := loops
	// if loops == -1 {
	// 	numLoops = -1
	// } else if loops == 0 {
	// 	numLoops = 0
	// } else {
	// 	numLoops = loops
	// }

	loopStreamer := beep.Loop(numLoops, streamer)

	// Resample the audio track to 44100 Hz
	resampled := beep.Resample(4, format.SampleRate, 44100, loopStreamer)

	// Adjust the audio for the volume.
	// The volume value should be a ratio for the volume adjustment.
	// if volume is 0, then mute the audio
	volumeStreamer := &effects.Volume{
		Streamer: resampled,
		Base:     1.5,
		Volume:   volume - 5,
		Silent:   volume == 0,
	}

	// Allow for the speed of the track to be changed.
	// The speed value should be a ratio for the speed adjustment. A value of 1 is the original speed.
	speedStreamer := beep.ResampleRatio(4, speed, volumeStreamer)

	// Allow for playing and pausing the audio track
	ctrl := &beep.Ctrl{Streamer: speedStreamer, Paused: false}

	s.state = "playing"

	// Play the audio track.
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		s.done <- true
	})))

	for {
		select {
		case <-s.done:
			return nil
		case <-s.pauseChannel:
			speaker.Lock()
			ctrl.Paused = true
			speaker.Unlock()
			s.state = "paused"
		case <-s.resumeChannel:
			speaker.Lock()
			ctrl.Paused = false
			speaker.Unlock()
			s.state = "playing"
		case <-s.clear:
			speaker.Lock()
			speaker.Clear()
			speaker.Unlock()
			s.state = "cleared"
			return nil
		case v := <-s.volumeChange:
			speaker.Lock()
			volumeStreamer.Volume = v - 5
			speaker.Unlock()
		case volumeDiff := <-s.masterVolumeChannel:
			speaker.Lock()
			volumeStreamer.Volume = volumeStreamer.Volume + volumeDiff
			speaker.Unlock()
		case <-s.muteChannel:
			speaker.Lock()
			volumeStreamer.Silent = true
			speaker.Unlock()
		case <-s.unmuteChannel:
			speaker.Lock()
			volumeStreamer.Silent = false
			speaker.Unlock()
			// case <-time.After(time.Second):
			//    speaker.Lock()
			//    fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
			//    speaker.Unlock()
		}
	}
}

func StopAudio() {
	for _, track := range SpeakerTracks {
		if track != nil {
			track.clear <- true
		}
	}
	// remove all tracks
	// SpeakerTracks = make(map[string]*SpeakerTrack)
}

// Pause the audio track.
func (s *SpeakerTrack) PauseAudio() {
	s.pauseChannel <- true
}

// Resume the audio track.
func (s *SpeakerTrack) ResumeAudio() {
	s.resumeChannel <- true
}

// Change the volume of the audio track.
func (s *SpeakerTrack) ChangeVolume(volume float64) {
	s.volumeChange <- volume
}

// Changes master volume of all audio tracks. The default master volume is 5.
func ChangeMasterVolume(volume float64) {
	if volume <= 0 {
		volume = 0
	} else if volume > 10 {
		volume = 10
	}
	volumeDiff := volume - float64(masterVolume)

	masterVolume = volume

	for _, track := range SpeakerTracks {
		if track != nil {
			track.masterVolumeChannel <- volumeDiff
			if volume == 0 {
				track.muteChannel <- true
			} else {
				track.unmuteChannel <- true
			}
		}
	}

}

func (s *SpeakerTrack) GetName() string {
	return s.name
}
