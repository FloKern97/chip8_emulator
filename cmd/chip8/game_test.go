package main

import (
	"testing"

	"my_emulator/internal/chip8"
)

func TestLoadROMIntoGame(t *testing.T) {
	game := &Game{}
	rom := []byte{0x12, 0x34}

	if err := game.loadROM(rom, "test.ch8"); err != nil {
		t.Fatalf("unexpected error loading ROM: %v", err)
	}

	if !game.romLoaded {
		t.Fatal("expected game to be marked as having a loaded ROM")
	}
	if game.romName != "test.ch8" {
		t.Fatalf("expected ROM name test.ch8, got %q", game.romName)
	}
	if game.gameCpu.PC != 0x200 {
		t.Fatalf("expected new ROM CPU to start at 0x200, got 0x%X", game.gameCpu.PC)
	}
	if game.gameCpu.Memory[0x200] != 0x12 || game.gameCpu.Memory[0x201] != 0x34 {
		t.Fatalf("expected ROM bytes to be loaded at 0x200, got %02X %02X",
			game.gameCpu.Memory[0x200], game.gameCpu.Memory[0x201])
	}
}

func TestLoadOversizedROMKeepsCurrentGame(t *testing.T) {
	currentCPU := chip8.NewCPU()
	game := &Game{
		gameCpu:   currentCPU,
		romLoaded: true,
		romName:   "current.ch8",
	}

	err := game.loadROM(make([]byte, 4096-0x200+1), "too-large.ch8")
	if err == nil {
		t.Fatal("expected an error for an oversized ROM")
	}
	if game.gameCpu != currentCPU {
		t.Fatal("expected the current CPU to remain active after a failed ROM load")
	}
	if game.romName != "current.ch8" {
		t.Fatalf("expected current ROM name to remain unchanged, got %q", game.romName)
	}
}
