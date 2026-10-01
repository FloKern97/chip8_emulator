package main

import (
	"image"
	"my_emulator/internal/chip8"
	"github.com/hajimehoshi/ebiten/v2"
)

const COMMANDS_PER_LOOP = 10

type Game struct {
	gameCpu *chip8.CPU
}

func (g *Game) Update() error {
	// Update game loop
	for range COMMANDS_PER_LOOP {
		_, err := g.gameCpu.FetchAndExecuteCommand()

		if err != nil {
			return err
		}
	}
	
	g.gameCpu.DecreaseDelayAndSoundTimer()


	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(image.Black)
	for y := 0; y < len(g.gameCpu.Display); y++ {
		for x := 0; x < len(g.gameCpu.Display[y]); x++ {
			if g.gameCpu.Display[y][x] {
				screen.Set(x, y, image.White)
			}
		}
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {

	return 64, 32
}
