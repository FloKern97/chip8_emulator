package chip8

import "errors"
import "math/rand"

const (
	memorySize   = 4096
	programStart = 0x200
	registers    = 16
	stackSize    = 16
	byteSize     = 8
)

type CPU struct {
	Memory  [memorySize]byte
	V       [registers]byte
	I       uint16
	PC      uint16
	Stack   [stackSize]uint16
	SP      byte
	Delay   byte
	Sound   byte
	Display [32][64]bool
	Keys    [16]bool
}

func NewCPU() *CPU {
	return &CPU{
		PC: programStart,
	}
}

func (cpu *CPU) LoadROM(rom []byte) error {
	if len(rom) > len(cpu.Memory)-programStart {
		return errors.New("rom too large for cpu memory")
	}

	for index, value := range rom {
		cpu.Memory[programStart+index] = value
	}

	return nil
}

func (cpu *CPU) Fetch() uint16 {
	high := cpu.Memory[cpu.PC]
	low := cpu.Memory[cpu.PC+1]

	shiftedHigh := uint16(high) << 8

	opCode := shiftedHigh | uint16(low)

	cpu.PC += 2

	return uint16(opCode)
}

func (cpu *CPU) Execute(opCode uint16) error {
	switch opCode & 0xF000 {
	// 0NNN/00E0/00EE: special opcodes; inspect the full opcode in handleSpecialCodes.
	case 0x0000:
		return cpu.handleSpecialCodes(opCode)

	// 1NNN: jump to address NNN.
	case 0x1000:
		cpu.PC = opCode & 0x0FFF

	// 2NNN: call subroutine at NNN and store the return address on the stack.
	case 0x2000:
		subRoutineAddress := opCode & 0x0FFF
		cpu.Stack[cpu.SP] = cpu.PC
		cpu.SP++

		cpu.PC = subRoutineAddress

	// 3XNN: skip next instruction if VX equals NN.
	case 0x3000:
		register := cpu.readFirstRegisterFromOpCode(opCode)
		value := opCode & 0x00FF

		if cpu.V[register] == byte(value) {
			cpu.PC += 2
		}

	// 4XNN: skip next instruction if VX does not equal NN.
	case 0x4000:
		register := cpu.readFirstRegisterFromOpCode(opCode)
		value := opCode & 0x00FF

		if cpu.V[register] != byte(value) {
			cpu.PC += 2
		}

	// 5XY0: skip next instruction if VX equals VY.
	case 0x5000:
		firstRegister := cpu.readFirstRegisterFromOpCode(opCode)
		secondRegister := cpu.readSecondRegisterFromOpCode(opCode)
		lastBit := opCode & 0x000F

		if lastBit != 0 {
			return errors.New("last bit of the opcode must be 0 for this operation (e.g. 0x5AB0)")
		}

		if cpu.V[firstRegister] == cpu.V[secondRegister] {
			cpu.PC += 2
		}

	// 6XNN: set VX to NN.
	case 0x6000:
		register := cpu.readFirstRegisterFromOpCode(opCode)
		registerValue := opCode & 0x00FF
		cpu.V[register] = byte(registerValue)

	// 7XNN: add NN to VX without changing VF.
	case 0x7000:
		register := cpu.readFirstRegisterFromOpCode(opCode)
		valueToAdd := opCode & 0x00FF
		cpu.V[register] += byte(valueToAdd)

	// 8XYN: register operation selected by the final nibble N.
	case 0x8000:
		firstRegister := cpu.readFirstRegisterFromOpCode(opCode)
		secondRegister := cpu.readSecondRegisterFromOpCode(opCode)
		lastBitCommand := int(opCode & 0x000F)

		return cpu.handleSubCommand(firstRegister, secondRegister, lastBitCommand)

	// 9XY0: skip next instruction if VX does not equal VY.
	case 0x9000:
		lastBit := opCode & 0x000F

		if lastBit != 0 {
			return errors.New("The last bit must be zero for this operation (e.g. 0x9AB0)")
		}

		firstRegister := cpu.readFirstRegisterFromOpCode(opCode)
		secondRegister := cpu.readSecondRegisterFromOpCode(opCode)

		if cpu.V[firstRegister] != cpu.V[secondRegister] {
			cpu.PC += 2
		}
	// ANNN: set I to address NNN.
	case 0xA000:
		cpu.I = opCode & 0x0FFF

	// BNNN: jump to address NNN plus V0.
	case 0xB000:
		opCodeValue := opCode & 0x0FFF
		cpu.PC = opCodeValue + uint16(cpu.V[0])

	// CXNN: set VX to a random byte masked with NN.
	case 0xC000:
		randomNumber := rand.Int()
		commandValue := opCode & 0x00FF
		registerToSet := cpu.readFirstRegisterFromOpCode(opCode)

		cpu.V[registerToSet] = byte(randomNumber) & byte(commandValue)

	// DXYN: draw an N-byte sprite at VX,VY and set VF on collision.
	case 0xD000:
		cpu.handleDisplayDraw(opCode)

	// EX9E and EXA1: Skip next command if button is pressed (EX9E) or if it is not pressed (EXA1)
	case 0xE000:
		if opCode&0x00FF == 0xA1 {
			cpu.handleSkipForUnpressedButton(opCode)
		}

		if opCode&0x00FF == 0x9E {
			cpu.handleSkipForPressedButton(opCode)
		}

	case 0xF000:
		return cpu.handleTimerCodes(opCode)

	default:
		return errors.New("unsupported opcode")
	}

	return nil
}

func (cpu *CPU) handleSpecialCodes(opCode uint16) error {
	switch opCode {
	// 00EE: return from the current subroutine.
	case 0x00EE:
		cpu.SP--
		cpu.PC = cpu.Stack[cpu.SP]

	// 00E0: clear the display.
	case 0x00E0:
		for y := 0; y < 32; y++ {
			for x := 0; x < 64; x++ {
				cpu.Display[y][x] = false
			}
		}

	default:
		return errors.New("unsupported special code")
	}

	return nil
}

func (cpu *CPU) handleDisplayDraw(opCode uint16) {
	cpu.V[15] = 0
	xCoordinate := cpu.V[int((opCode&0x0F00)>>8)]
	yCoordinate := cpu.V[int((opCode&0x00F0)>>4)]
	height := opCode & 0x000F

	for spriteRow := 0; spriteRow < int(height); spriteRow++ {
		spriteByte := cpu.Memory[int(cpu.I)+spriteRow]

		for pixelColumn := 0; pixelColumn < byteSize; pixelColumn++ {
			mask := byte(0x80 >> pixelColumn)

			if spriteByte&mask != 0 {
				wasActivePixel := cpu.Display[int(yCoordinate)+spriteRow][int(xCoordinate)+pixelColumn]
				cpu.Display[int(yCoordinate)+spriteRow][int(xCoordinate)+pixelColumn] = !cpu.Display[int(yCoordinate)+spriteRow][int(xCoordinate)+pixelColumn]

				if wasActivePixel {
					cpu.V[15] = 1
				}
			}
		}
	}
}

func (cpu *CPU) handleSubCommand(firstRegister int, secondRegister int, lastBitCommand int) error {
	switch lastBitCommand {
	// 8XY0: set VX to the value of VY.
	case 0:
		cpu.V[firstRegister] = cpu.V[secondRegister]

	// 8XY1: set VX to VX OR VY.
	case 1:
		cpu.V[firstRegister] = cpu.V[firstRegister] | cpu.V[secondRegister]

	// 8XY2: set VX to VX AND VY.
	case 2:
		cpu.V[firstRegister] = cpu.V[firstRegister] & cpu.V[secondRegister]

	// 8XY3: set VX to VX XOR VY.
	case 3:
		cpu.V[firstRegister] = cpu.V[firstRegister] ^ cpu.V[secondRegister]

	// 8XY4: add VY to VX and set VF if the result overflows one byte.
	case 4:
		firstRegisterValue := uint16(cpu.V[firstRegister])
		secondRegisterValue := uint16(cpu.V[secondRegister])
		sum := firstRegisterValue + secondRegisterValue

		cpu.V[firstRegister] = byte(sum)

		if sum > 0xFF {
			cpu.V[15] = 1
		} else {
			cpu.V[15] = 0
		}

	// 8XY5: subtract VY from VX and set VF to 0 if a borrow is needed.
	case 5:
		firstRegisterValue := uint16(cpu.V[firstRegister])
		secondRegisterValue := uint16(cpu.V[secondRegister])
		cpu.V[firstRegister] = byte(firstRegisterValue - secondRegisterValue)

		if firstRegisterValue < secondRegisterValue {
			cpu.V[15] = 0
		} else {
			cpu.V[15] = 1
		}

	// 8XY6: shift VX right once and store the old lowest bit in VF.
	case 6:
		firstRegisterValue := uint16(cpu.V[firstRegister])
		if firstRegisterValue&0x01 == 0x01 {
			cpu.V[15] = 1

		} else {
			cpu.V[15] = 0
		}
		firstRegisterValue = firstRegisterValue >> 1
		cpu.V[firstRegister] = byte(firstRegisterValue)

	// 8XY7: set VX to VY minus VX and set VF to 0 if a borrow is needed.
	case 7:
		firstRegisterValue := uint16(cpu.V[firstRegister])
		secondRegisterValue := uint16(cpu.V[secondRegister])
		cpu.V[firstRegister] = byte(secondRegisterValue - firstRegisterValue)

		if secondRegisterValue < firstRegisterValue {
			cpu.V[15] = 0
		} else {
			cpu.V[15] = 1
		}

	// 8XYE: shift VX left once and store the old highest bit in VF.
	case 14:
		firstRegisterValue := uint16(cpu.V[firstRegister])
		if firstRegisterValue&0x80 == 0x80 {
			cpu.V[15] = 1

		} else {
			cpu.V[15] = 0
		}
		firstRegisterValue = firstRegisterValue << 1
		cpu.V[firstRegister] = byte(firstRegisterValue)
	default:
		return errors.New("unsupported subcommand")
	}

	return nil
}

func (cpu *CPU) handleSkipForPressedButton(opCode uint16) {
	register := cpu.readFirstRegisterFromOpCode(opCode)
	key := cpu.V[register]

	if cpu.Keys[int(key)] {
		cpu.PC += 2
	}
}

func (cpu *CPU) handleSkipForUnpressedButton(opCode uint16) {
	register := cpu.readFirstRegisterFromOpCode(opCode)
	key := cpu.V[register]

	if !cpu.Keys[int(key)] {
		cpu.PC += 2
	}
}

func (cpu *CPU) handleTimerCodes(opCode uint16) error {
	switch opCode & 0x00FF {

	// Reading from Delay Timer: VX = Delay Timer
	case 0x07:
		registerToWrite := cpu.readFirstRegisterFromOpCode(opCode)
		cpu.V[registerToWrite] = cpu.Delay

	// Setting Delay Timer: Delay = VX
	case 0x15:
		registerToRead := cpu.readFirstRegisterFromOpCode(opCode)
		cpu.Delay = cpu.V[registerToRead]

	// Setting sound: cpu.sound = VX
	case 0x18:
		registerToRead := cpu.readFirstRegisterFromOpCode(opCode)
		cpu.Sound = cpu.V[registerToRead]

	case 0x1E:
		registerToRead := cpu.readFirstRegisterFromOpCode(opCode)

		cpu.I += uint16(cpu.V[registerToRead])

	case 0x55:
		register := cpu.readFirstRegisterFromOpCode(opCode)

		for index := 0; index <= register; index++ {
			memoryIndex := cpu.I + uint16(index)
			cpu.Memory[memoryIndex] = cpu.V[index]
		}

	case 0x65:
		register := cpu.readFirstRegisterFromOpCode(opCode)

		for index := 0; index <= register; index++ {
			memoryIndex := cpu.I + uint16(index)
			cpu.V[index] = cpu.Memory[memoryIndex]
		}

	default:
		return errors.New("unsupported timer opcode")
	}

	return nil
}

func (cpu *CPU) readFirstRegisterFromOpCode(opCode uint16) int {
	return int((opCode & 0x0F00) >> 8)
}

func (cpu *CPU) readSecondRegisterFromOpCode(opCode uint16) int {
	return int((opCode & 0x00F0) >> 4)
}
