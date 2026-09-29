package main

import (
	"fmt"
	"my_emulator/internal/chip8"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Fehler: Es wurde scheinbar kein ROM-Pfad angegeben. Eingabe bitte überprüfen.")
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

	for i := 0; i < 2; i++ {
		
		opcode, executeErr := cpu.FetchAndExecuteCommand()

		if executeErr != nil {
			fmt.Fprintf(os.Stderr, "An error occurred while executing opcode %X: %v.\n", opcode, executeErr)
			return
		}

		fmt.Printf("Executed opcode: %X\n", opcode)
		fmt.Printf("Current Program Count: %X\n", cpu.PC)
	}
}
