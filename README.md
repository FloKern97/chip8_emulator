# My Emulator

This repository is a standalone learning project for building a CHIP-8 emulator in Go.

It covers core emulator concepts:

- CPU state
- memory maps
- fetch/decode/execute loops
- opcodes
- timers
- display output
- input handling
- tests for machine behavior

This project is focused on CHIP-8. Other emulator projects, such as a future
Game Boy Advance emulator, will live in separate repositories.

## Commands

Run all tests:

```sh
go test ./...
```

Run the CHIP-8 emulator, then select a ROM using the **Open ROM** button:

```sh
go run ./cmd/chip8
```

The emulator opens a window with a file picker. Choose a CHIP-8 ROM to start
playing. ROM files without a file extension are supported; the picker shows
all files.

### Keyboard

The PC keyboard maps to the CHIP-8 keypad as follows:

```text
CHIP-8:  1 2 3 C       PC:  1 2 3 4
         4 5 6 D            Q W E R
         7 8 9 E            A S D F
         A 0 B F            Z X C V
```

WASD are also movement aliases for the common CHIP-8 direction keys:
W = 2 (up), A = 4 (left), S = 8 (down), and D = 6 (right). The original
mapping remains active as well.

Press **P** to pause or resume the currently loaded ROM.

Select **Controls** in the window's navbar to view the original CHIP-8 keypad
mapping and the extra WASD movement aliases. The game pauses while this panel
is open; close it with **Close (Esc)** or the Escape key.

### Build

Build a standalone executable:

```sh
go build -o chip8.exe ./cmd/chip8
```

Then start the emulator:

```powershell
.\chip8.exe
```

Choose a ROM from the window's **Open ROM** button.
