package main

import (
	"fmt"
	"my_emulator/internal/chip8"
	"os"
	"strconv"
	"time"
)

func main() {
	const COMMANDS_PER_LOOP = 10

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: please provide a ROM-path and a loop counter.")
		return
	}

	romPath := os.Args[1]
	commandLoops, loopCounterError := strconv.Atoi(os.Args[2])

	if loopCounterError != nil {
		fmt.Fprintf(os.Stderr, "Error while reading loop counter: %v", loopCounterError)
		return
	}

	if commandLoops < 1 {
		fmt.Fprintf(os.Stderr, "Loop counter argument must be a positive integer.")
		return
	}

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

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	remainingCommands := commandLoops

	for remainingCommands > 0 {
		<-ticker.C
		commandsToExecute := remainingCommands

		if commandsToExecute > COMMANDS_PER_LOOP {
			commandsToExecute = COMMANDS_PER_LOOP
		}
		
		for range commandsToExecute {
		opcode, executeErr := cpu.FetchAndExecuteCommand()

		if executeErr != nil {
			fmt.Fprintf(os.Stderr, "An error occurred while executing opcode %X: %v.\n", opcode, executeErr)
			return
		}

		fmt.Printf("Executed opcode: %X\n", opcode)
		fmt.Printf("Current Program Count: %X\n", cpu.PC)

		}
		remainingCommands -= commandsToExecute
		cpu.DecreaseDelayAndSoundTimer()
	}
}
