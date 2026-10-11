package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const audioSampleRate = 44100

func main() {
	// TODO Maybe add dynamic window size some other time
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	audioContext := audio.NewContext(audioSampleRate)
	soundPlayer, err := newSoundPlayer(audioContext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CHIP-8 sound player: %v\n", err)
		return
	}

	gameError := ebiten.RunGame(&Game{
		soundPlayer: soundPlayer,
		status:      "Select a ROM to start",
	})

	if gameError != nil {
		fmt.Fprintf(os.Stderr, "Error running CHIP-8 window: %v\n", gameError)
	}
}
