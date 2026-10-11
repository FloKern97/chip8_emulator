package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"my_emulator/internal/chip8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sqweek/dialog"
)

const (
	commandsPerUpdate   = 10
	menuBarHeight       = 40
	displayScale        = 10
	openButtonX         = 8
	openButtonY         = 8
	openButtonWidth     = 104
	openButtonHeight    = 24
	controlsButtonX     = 120
	controlsButtonWidth = 100
	controlsPanelX      = 70
	controlsPanelY      = 74
	controlsPanelWidth  = 500
	controlsPanelHeight = 252
	closeButtonX        = 450
	closeButtonY        = 286
	closeButtonWidth    = 100
	closeButtonHeight   = 26
)

var keyMapping = [16]ebiten.Key{
	0x0: ebiten.KeyX,
	0x1: ebiten.Key1,
	0x2: ebiten.Key2,
	0x3: ebiten.Key3,
	0x4: ebiten.KeyQ,
	0x5: ebiten.KeyW,
	0x6: ebiten.KeyE,
	0x7: ebiten.KeyA,
	0x8: ebiten.KeyS,
	0x9: ebiten.KeyD,
	0xA: ebiten.KeyZ,
	0xB: ebiten.KeyC,
	0xC: ebiten.Key4,
	0xD: ebiten.KeyR,
	0xE: ebiten.KeyF,
	0xF: ebiten.KeyV,
}

type Game struct {
	gameCpu      *chip8.CPU
	soundPlayer  *audio.Player
	romLoaded    bool
	paused       bool
	controlsOpen bool
	romName      string
	status       string
}

func newSoundPlayer(audioContext *audio.Context) (*audio.Player, error) {
	const samplesPerWave = 100
	const amplitude = 5000
	pcm := make([]byte, samplesPerWave*4)

	for sample := 0; sample < samplesPerWave; sample++ {
		value := int16(-amplitude)
		if sample < samplesPerWave/2 {
			value = amplitude
		}

		offset := sample * 4
		binary.LittleEndian.PutUint16(pcm[offset:offset+2], uint16(value))
		binary.LittleEndian.PutUint16(pcm[offset+2:offset+4], uint16(value))
	}

	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	return audioContext.NewPlayer(loop)
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()
	isHoveringOpenButton := x >= openButtonX && x < openButtonX+openButtonWidth &&
		y >= openButtonY && y < openButtonY+openButtonHeight
	isHoveringControlsButton := x >= controlsButtonX && x < controlsButtonX+controlsButtonWidth &&
		y >= openButtonY && y < openButtonY+openButtonHeight
	isHoveringCloseButton := x >= closeButtonX && x < closeButtonX+closeButtonWidth &&
		y >= closeButtonY && y < closeButtonY+closeButtonHeight

	if isHoveringOpenButton || isHoveringControlsButton || (g.controlsOpen && isHoveringCloseButton) {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.controlsOpen {
			if isHoveringCloseButton {
				g.closeControls()
			}
		} else {
			if isHoveringOpenButton {
				g.openROM()
			}
			if isHoveringControlsButton {
				g.openControls()
			}
		}
	}

	if g.controlsOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.closeControls()
		}
		return nil
	}

	if g.romLoaded && inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
		if g.paused && g.soundPlayer.IsPlaying() {
			g.soundPlayer.Pause()
		}
	}

	if !g.romLoaded || g.paused {
		return nil
	}

	for chip8Key, keyboardKey := range keyMapping {
		g.gameCpu.Keys[chip8Key] = ebiten.IsKeyPressed(keyboardKey)
	}

	g.gameCpu.Keys[0x2] = g.gameCpu.Keys[0x2] || ebiten.IsKeyPressed(ebiten.KeyW)
	g.gameCpu.Keys[0x4] = g.gameCpu.Keys[0x4] || ebiten.IsKeyPressed(ebiten.KeyA)
	g.gameCpu.Keys[0x6] = g.gameCpu.Keys[0x6] || ebiten.IsKeyPressed(ebiten.KeyD)
	g.gameCpu.Keys[0x8] = g.gameCpu.Keys[0x8] || ebiten.IsKeyPressed(ebiten.KeyS)

	for range commandsPerUpdate {
		_, err := g.gameCpu.FetchAndExecuteCommand()

		if err != nil {
			return err
		}
	}

	g.gameCpu.DecreaseDelayAndSoundTimer()

	if g.gameCpu.Sound > 0 {
		if !g.soundPlayer.IsPlaying() {
			g.soundPlayer.Play()
		}
	} else if g.soundPlayer.IsPlaying() {
		g.soundPlayer.Pause()
	}

	return nil
}

func (g *Game) openControls() {
	g.controlsOpen = true
	if g.soundPlayer != nil && g.soundPlayer.IsPlaying() {
		g.soundPlayer.Pause()
	}
}

func (g *Game) closeControls() {
	g.controlsOpen = false
}

func (g *Game) openROM() {
	filename, err := dialog.File().
		Title("Open CHIP-8 ROM").
		Load()
	if err == dialog.ErrCancelled {
		return
	}
	if err != nil {
		g.status = fmt.Sprintf("Could not open file picker: %v", err)
		return
	}

	rom, err := os.ReadFile(filename)
	if err != nil {
		g.status = fmt.Sprintf("Could not read ROM: %v", err)
		return
	}
	if err := g.loadROM(rom, filepath.Base(filename)); err != nil {
		g.status = err.Error()
		return
	}

	g.status = ""
}

func (g *Game) loadROM(rom []byte, name string) error {
	cpu := chip8.NewCPU()
	if err := cpu.LoadROM(rom); err != nil {
		return fmt.Errorf("could not load ROM: %w", err)
	}

	g.gameCpu = cpu
	g.romLoaded = true
	g.romName = name
	if g.soundPlayer != nil && g.soundPlayer.IsPlaying() {
		g.soundPlayer.Pause()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(image.Black)
	vector.FillRect(screen, 0, 0, 640, float32(menuBarHeight), color.RGBA{R: 40, G: 40, B: 40, A: 255}, false)
	cursorX, cursorY := ebiten.CursorPosition()
	openButtonColor := color.RGBA{R: 70, G: 110, B: 170, A: 255}
	if cursorX >= openButtonX && cursorX < openButtonX+openButtonWidth &&
		cursorY >= openButtonY && cursorY < openButtonY+openButtonHeight {
		openButtonColor = color.RGBA{R: 90, G: 140, B: 210, A: 255}
	}
	vector.FillRect(screen, float32(openButtonX), float32(openButtonY), float32(openButtonWidth), float32(openButtonHeight), openButtonColor, false)
	ebitenutil.DebugPrintAt(screen, "Open ROM", openButtonX+12, openButtonY+8)
	controlsButtonColor := color.RGBA{R: 70, G: 110, B: 170, A: 255}
	if cursorX >= controlsButtonX && cursorX < controlsButtonX+controlsButtonWidth &&
		cursorY >= openButtonY && cursorY < openButtonY+openButtonHeight {
		controlsButtonColor = color.RGBA{R: 90, G: 140, B: 210, A: 255}
	}
	vector.FillRect(screen, float32(controlsButtonX), float32(openButtonY), float32(controlsButtonWidth), float32(openButtonHeight), controlsButtonColor, false)
	ebitenutil.DebugPrintAt(screen, "Controls", controlsButtonX+14, openButtonY+8)

	status := g.status
	if status == "" {
		if g.romLoaded {
			status = "ROM: " + g.romName
			if g.paused {
				status += "  [PAUSED]"
			}
		} else {
			status = "Select a ROM to start"
		}
	}
	ebitenutil.DebugPrintAt(screen, status, controlsButtonX+controlsButtonWidth+16, openButtonY+8)

	if g.romLoaded {
		for y := 0; y < len(g.gameCpu.Display); y++ {
			for x := 0; x < len(g.gameCpu.Display[y]); x++ {
				if g.gameCpu.Display[y][x] {
					vector.FillRect(
						screen,
						float32(x*displayScale),
						float32(menuBarHeight+y*displayScale),
						float32(displayScale),
						float32(displayScale),
						image.White,
						false,
					)
				}
			}
		}
	}

	if g.paused {
		vector.FillRect(screen, 180, 160, 280, 40, color.RGBA{R: 0, G: 0, B: 0, A: 220}, false)
		ebitenutil.DebugPrintAt(screen, "PAUSED - press P to resume", 196, 176)
	}

	if g.controlsOpen {
		vector.FillRect(screen, 0, float32(menuBarHeight), 640, 320, color.RGBA{A: 210}, false)
		vector.FillRect(screen, controlsPanelX, controlsPanelY, controlsPanelWidth, controlsPanelHeight, color.RGBA{R: 32, G: 36, B: 44, A: 255}, false)
		ebitenutil.DebugPrintAt(screen, "CHIP-8 CONTROLS", controlsPanelX+20, controlsPanelY+18)
		ebitenutil.DebugPrintAt(screen, "Original CHIP-8 keypad       Keyboard", controlsPanelX+20, controlsPanelY+44)
		ebitenutil.DebugPrintAt(screen, "1 2 3 C                      1 2 3 4", controlsPanelX+20, controlsPanelY+66)
		ebitenutil.DebugPrintAt(screen, "4 5 6 D                      Q W E R", controlsPanelX+20, controlsPanelY+84)
		ebitenutil.DebugPrintAt(screen, "7 8 9 E                      A S D F", controlsPanelX+20, controlsPanelY+102)
		ebitenutil.DebugPrintAt(screen, "A 0 B F                      Z X C V", controlsPanelX+20, controlsPanelY+120)
		ebitenutil.DebugPrintAt(screen, "Extra movement keys: WASD", controlsPanelX+20, controlsPanelY+150)
		ebitenutil.DebugPrintAt(screen, "W = 2 (up), A = 4 (left)", controlsPanelX+20, controlsPanelY+170)
		ebitenutil.DebugPrintAt(screen, "S = 8 (down), D = 6 (right)", controlsPanelX+20, controlsPanelY+188)
		ebitenutil.DebugPrintAt(screen, "P = pause / resume", controlsPanelX+20, controlsPanelY+214)
		closeButtonColor := color.RGBA{R: 70, G: 110, B: 170, A: 255}
		if cursorX >= closeButtonX && cursorX < closeButtonX+closeButtonWidth &&
			cursorY >= closeButtonY && cursorY < closeButtonY+closeButtonHeight {
			closeButtonColor = color.RGBA{R: 90, G: 140, B: 210, A: 255}
		}
		vector.FillRect(screen, closeButtonX, closeButtonY, closeButtonWidth, closeButtonHeight, closeButtonColor, false)
		ebitenutil.DebugPrintAt(screen, "Close (Esc)", closeButtonX+12, closeButtonY+8)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, menuBarHeight + 320
}
