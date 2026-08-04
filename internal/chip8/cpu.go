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
	// jump
	case 0x1000:
		cpu.PC = opCode & 0x0FFF

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
		
		// Set I
	case 0xA000:
		cpu.I = opCode & 0x0FFF
	default:
		return errors.New("unsupported opcode")
	}

	return nil
}
