package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"my_emulator/internal/chip8"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: please provide a ROM path.")
		return
	}

	romPath := os.Args[1]

	rom, err := os.ReadFile(romPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading the ROM file: %v", err)
		return
	}

	cpu := chip8.NewCPU()
	loadingErr := cpu.LoadROM(rom)

	if loadingErr != nil {
		fmt.Fprintf(os.Stderr, "An error occurred while processing ROM file: %v", loadingErr)
		return
	}

	fmt.Println("Successfully read ROM file into memory.")
	ebiten.SetWindowSize(640, 320)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	gameError := ebiten.RunGame(&Game{gameCpu: cpu})

	if gameError != nil {
		fmt.Fprintf(os.Stderr, "Error running CHIP-8 window: %v\n", gameError)
	}
}
