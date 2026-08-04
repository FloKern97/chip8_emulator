package chip8

import "testing"

func TestNewCPU(t *testing.T) {
	cpu := NewCPU()

	if cpu.PC != 0x200 {
		t.Fatalf("expected PC to start at 0x200 got 0x%X", cpu.PC)
	}

	if cpu.I != 0 {
		t.Fatalf("expected I to start at 0 got 0x%X", cpu.I)
	}

	if cpu.Delay != 0 {
		t.Fatalf("expected Delay to start at 0 got 0x%X", cpu.Delay)
	}

	if cpu.Sound != 0 {
		t.Fatalf("expected Sound to start at 0 got 0x%X", cpu.Sound)
	}

	if cpu.SP != 0 {
		t.Fatalf("expected SP to start at 0 got 0x%X", cpu.SP)
	}
}

func TestLoadROM(t *testing.T) {
	cpu := NewCPU()

	rom := []byte{0x14, 0x36, 0xFF, 0xED}

	err := cpu.LoadROM(rom)

	if err != nil {
		t.Fatalf("expected LoadROM to succeed got error: %v", err)
	}

	if cpu.Memory[0x200] != 0x14 {
		t.Fatalf("expected Memory[0x200] to be 0x14 got 0x%X", cpu.Memory[0x200])
	}

	if cpu.Memory[0x201] != 0x36 {
		t.Fatalf("expected Memory[0x201] to be 0x36 got 0x%X", cpu.Memory[0x201])
	}

	if cpu.Memory[0x202] != 0xFF {
		t.Fatalf("expected Memory[0x202] to be 0xFF got 0x%X", cpu.Memory[0x202])
	}

	if cpu.Memory[0x203] != 0xED {
		t.Fatalf("expected Memory[0x203] to be 0xED got 0x%X", cpu.Memory[0x203])
	}
}

func TestLoadROM_Overflow(t *testing.T) {
	cpu := NewCPU()

	rom := make([]byte, 3885)

	err := cpu.LoadROM(rom)

	if err == nil {
		t.Fatal("expected error for ROM that is too large.")
	}
}

func TestFetchUInt16OpCode(t *testing.T) {
	cpu := NewCPU()

	rom := []byte{0x12, 0x34}

	err := cpu.LoadROM(rom)

	if err != nil {
		t.Fatal("an error occurred while loading the ROM.")
	}

	opCode := cpu.Fetch()

	if opCode != 0x1234 {
		t.Fatal("Fetch() did not calculate the correct result")
	}

	if cpu.PC != 0x202 {
		t.Fatal("pc was not set correctly after fetching the opCode.")
	}
}

func TestExecuteJump(t *testing.T) {
	cpu := NewCPU()

	jumpErr := cpu.Execute(0x1234)

	if jumpErr != nil {
		t.Fatal("an error occurred while executing the jump.")
	}

	if cpu.PC != 0x234 {
		t.Fatalf("the new PC address was calculated incorrectly. Expected 0x234, got 0x%X", cpu.PC)
	}
}

func TestExecuteUnknownOpCode(t *testing.T) {
	cpu := NewCPU()

	err := cpu.Execute(0xFFFF)

	if err == nil {
		t.Fatal("0xFFFF should be an unknown Opcode and cause an error.")
	}
}

func TestExecuteSetRegister(t *testing.T) {
	cpu := NewCPU()

	err := cpu.Execute(0x6A05)

	if err != nil {
		t.Fatal("an error occurred while trying to write into the register.")
	}

	if cpu.V[10] != 0x05 {
		t.Fatalf("expected V[10] to be 0x05, got 0x%X", cpu.V[10])
	}
}

func TestExecuteAddToRegister(t *testing.T) {
	cpu := NewCPU()
	cpu.V[9] = 0x05

	err := cpu.Execute(0x7913)

	if err != nil {
		t.Fatal("an error occurred while trying to write into the register.")
	}

	if cpu.V[9] != 0x18 {
		t.Fatalf("expected V[9] to be 0x18, got 0x%X", cpu.V[9])
	}
}

func TestExecuteWriteToI(t *testing.T) {
	cpu := NewCPU()

	err := cpu.Execute(0xA456)

	if err != nil {
		t.Fatal("an error occurred while trying to write into I.")
	}

	if cpu.I != 0x456 {
		t.Fatalf("expected cpu.I to be 0x456 got 0x%X", cpu.I)
	}
}

func TestExecuteSkip(t *testing.T) {
	cpu := NewCPU()
	cpu.PC = 0x202
	cpu.V[10] = 0x23

	err := cpu.Execute(0x3A23)

	if err != nil {
		t.Fatalf("an error occurred while processing an opcode")
	}

	if cpu.PC != 0x204 {
		t.Fatalf("cpu.PC has unexpected value, expected 0x204, got 0x%X", cpu.PC)
	}
}

func TestExecuteNoSkip(t *testing.T) {
	cpu := NewCPU()
	cpu.PC = 0x202
	cpu.V[10] = 0x23

	err := cpu.Execute(0x3A05)

	if err != nil {
		t.Fatalf("unexpected error while executing opcode: %v", err)
	}

	if cpu.PC != 0x202 {
		t.Fatalf("cpu.PC has unexpected value, expected 0x202, got 0x%X", cpu.PC)
	}
}

func TestExecuteSkipIfValueIsNotEqual(t *testing.T) {
	cpu := NewCPU()
	cpu.PC = 0x202
	cpu.V[10] = 0x23

	err := cpu.Execute(0x4A05)

	if err != nil {
		t.Fatalf("unexpected error while executing opcode: %v", err)
	}

	if cpu.PC != 0x204 {
		t.Fatalf("cpu.PC has unexpected value, expected 0x204, got 0x%X", cpu.PC)
	}
}

func TestExecuteNoSkipIfValueIsEqual(t *testing.T) {
	cpu := NewCPU()
	cpu.PC = 0x202
	cpu.V[10] = 0x05

	err := cpu.Execute(0x4A05)

	if err != nil {
		t.Fatalf("unexpected error while executing opcode: %v", err)
	}

	if cpu.PC != 0x202 {
		t.Fatalf("cpu.PC has unexpected value, expected 0x202, got 0x%X", cpu.PC)
	}
}

func TestExecuteNoSkipForMultipleRegisters(t *testing.T) {
	cpu := NewCPU()
	cpu.PC = 0x202

	cpu.V[9] = 0x34
	cpu.V[10] = 0x34

	err := cpu.Execute(0x59A0)

	if err != nil {
		t.Fatal("an error occurred while checking multiple registers for the same value")
	}

	if cpu.PC != 0x202 {
		t.Fatalf("expected PC to be 0x202 got 0x%X instead", cpu.PC)
	}
}

func TestExecuteSkipForMultipleRegisters(t *testing.T) {
	cpu := NewCPU()
	cpu.PC = 0x202

	cpu.V[9] = 0x33
	cpu.V[10] = 0x34

	err := cpu.Execute(0x59A0)

	if err != nil {
		t.Fatal("an error occurred while checking multiple registers for the same value")
	}

	if cpu.PC != 0x204 {
		t.Fatalf("expected PC to be 0x204 got 0x%X instead", cpu.PC)
	}
}
