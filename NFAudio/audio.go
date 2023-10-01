package NFAudio

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"os"
	"path/filepath"
	"time"
)

var SpeakerInitialized = false

func PlayAudio(file string, volume float64, looping bool) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}

	if !SpeakerInitialized {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			return err
		}
		SpeakerInitialized = true
	}

	go func() {
		defer streamer.Close()
		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))
		<-done
		log.Println("Finished playing audio file: " + filepath.Base(file))
	}()
	return nil
}
