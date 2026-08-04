package chip8

import "errors"

const (
	memorySize   = 4096
	programStart = 0x200
	registers    = 16
	stackSize    = 16
)

type CPU struct {
	Memory [memorySize]byte
	V      [registers]byte
	I      uint16
	PC     uint16
	Stack  [stackSize]uint16
	SP     byte
	Delay  byte
	Sound  byte
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
	// special op codes like returning from subroutine
	case 0x0000:
		return cpu.handleSpecialCodes(opCode)
	// jump
	case 0x1000:
		cpu.PC = opCode & 0x0FFF

	case 0x2000:
		subRoutineAddress := opCode & 0x0FFF
		cpu.Stack[cpu.SP] = cpu.PC
		cpu.SP++

		cpu.PC = subRoutineAddress

		// skip
	case 0x3000:
		register := int((opCode & 0x0F00) >> 8)
		value := opCode & 0x00FF

		if cpu.V[register] == byte(value) {
			cpu.PC += 2
		}

		// skip if value in cpu.V[N] is not equal to given value
	case 0x4000:
		register := int((opCode & 0x0F00) >> 8)
		value := opCode & 0x00FF

		if cpu.V[register] != byte(value) {
			cpu.PC += 2
		}

		// Compare register values, if not equal -> "skip"
	case 0x5000:
		firstRegister := int((opCode & 0x0F00) >> 8)
		secondRegister := int((opCode & 0x00F0) >> 4)
		lastBit := opCode & 0x000F

		if lastBit != 0 {
			return errors.New("last bit of the opcode must be 0 for this operation")
		}

		if cpu.V[firstRegister] == cpu.V[secondRegister] {
			cpu.PC += 2
		}

		// set register
	case 0x6000:
		register := int((opCode & 0x0F00) >> 8)
		registerValue := opCode & 0x00FF
		cpu.V[register] = byte(registerValue)

		// Add to register
	case 0x7000:
		register := int((opCode & 0x0F00) >> 8)
		valueToAdd := opCode & 0x00FF
		cpu.V[register] += byte(valueToAdd)

	case 0x8000:
		firstRegister := int((opCode & 0x0F00) >> 8)
		secondRegister := int((opCode & 0x0F0) >> 4)
		lastBitCommand := int(opCode & 0x000F)

		return cpu.handleSubCommand(firstRegister, secondRegister, lastBitCommand)

	// Set I
	case 0xA000:
		cpu.I = opCode & 0x0FFF

	default:
		return errors.New("unsupported opcode")
	}

	return nil
}

func (cpu *CPU) handleSpecialCodes(opCode uint16) error {
	switch opCode {
	// return from subroutine
	case 0x00EE:
		cpu.SP--
		cpu.PC = cpu.Stack[cpu.SP]

	default:
		return errors.New("unsupported special code")
	}

	return nil
}

func (cpu *CPU) handleSubCommand(firstRegister int, secondRegister int, lastBitCommand int) error {
	switch lastBitCommand {
	case 0:
		cpu.V[firstRegister] = cpu.V[secondRegister]
	
	// OR operation 
	case 1:
		cpu.V[firstRegister] = cpu.V[firstRegister] | cpu.V[secondRegister]
	
	// AND operation
	case 2:
		cpu.V[firstRegister] = cpu.V[firstRegister] & cpu.V[secondRegister]

	// XOR operation
	case 3:
		cpu.V[firstRegister] = cpu.V[firstRegister] ^ cpu.V[secondRegister]
	
	// Add operation
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
	// TODO add remaining cases 
	default:
		return errors.New("unsupported subcommand")
	}

	return nil
}
