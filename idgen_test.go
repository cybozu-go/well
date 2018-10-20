package well

import "testing"

func TestIDGenerator(t *testing.T) {
	t.Parallel()

	g := NewIDGenerator()
	if len(g.Generate()) != 36 {
		t.Error(len(g.Generate()) != 36)
	}

	g2 := &IDGenerator{
		[16]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x80, 0x90, 0xa0, 0xb0, 0xc0, 0xd0, 0xe0, 0xff},
		0xf0e0d0c001020304,
	}
	if g2.Generate() != "05020002-c4d5-e6f7-8090-a0b0c0d0e0ff" {
		t.Error(`g2.Generate() != "05020002-c4d5-e6f7-8090-a0b0c0d0e0ff"`)
	}
}

func BenchmarkIDGenerator(b *testing.B) {
	g := NewIDGenerator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Generate()
	}
}
