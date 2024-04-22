package main

import (
	"fmt"

	"go.novellaforge.dev/novellaforge/pkg/NFAudio"
)

func main() {

	// // Call the audio driver NFAudio
	// go NFAudio.PlayAudioFromFile("audio.mp3", 2, 1.0, false)

	// // wait for user input
	// fmt.Scanln()

	// // Change the volume of the audio
	// NFAudio.ChangeVolume(1)
	// fmt.Scanln()

	// // Pause the audio
	// NFAudio.PauseAudio()
	// fmt.Scanln()

	// // Resume the audio
	// NFAudio.ResumeAudio()
	// fmt.Scanln()

	// // Stop the audio
	// NFAudio.StopAudio()
	// fmt.Scanln()

	go NFAudio.PlayAudioFromFile("audio2.mp3", 3, 1, false)

	fmt.Scanln()

	go NFAudio.PlayAudioFromFile("audio.mp3", 2, 1, false)

	fmt.Scanln()

}
