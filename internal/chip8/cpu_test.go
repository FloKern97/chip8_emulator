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
		t.Fatal("An error occurred while loading the ROM.")
	}

	opCode := cpu.Fetch()

	if opCode != 0x1234 {
		t.Fatal("Fetch() did not calculate the correct result")
	}

	if cpu.PC != 0x202 {
		t.Fatal("PC was not set correctly after fetching the opCode.")
	}
}
