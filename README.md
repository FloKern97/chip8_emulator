# My Emulator

This repository is a learning project for building emulators in Go.

The first milestone is a CHIP-8 emulator, used to learn the core ideas:

- CPU state
- memory maps
- fetch/decode/execute loops
- opcodes
- timers
- display output
- input handling
- tests for machine behavior

After that, the project will move toward a Game Boy Advance emulator.

## Commands

Run all tests:

```sh
go test ./...
```

Run the CHIP-8 command:

```sh
go run ./cmd/chip8
```
