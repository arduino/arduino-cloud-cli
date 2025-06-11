package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleReceivedData_FullPacket(t *testing.T) {
	tc := NewTransportController()

	data := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55}
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55}
	got := tc.HandleReceivedData(data)

	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 1, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want, "expected packet bytes %v, got %v", want, got[0].ToBytes())
}

func TestHandleReceivedData_FullPacketWithJunk(t *testing.T) {
	tc := NewTransportController()

	data := []byte{0x65, 0x45, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55, 0x65, 0x24, 0x67}
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55}
	got := tc.HandleReceivedData(data)
	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 1, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want, "expected packet bytes %v, got %v", want, got[0].ToBytes())
}

func TestHandleReceivedData_SplitStartSequence(t *testing.T) {
	tc := NewTransportController()
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55}
	// First call: only 0x55
	got := tc.HandleReceivedData([]byte{0x55})

	assert.Nil(t, got, "expected nil packet list")

	// Second call: 0xaa and 0x02, should complete start and fill
	got = tc.HandleReceivedData([]byte{0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55})

	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 1, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want, "expected packet bytes %v, got %v", want, got[0].ToBytes())
}

func TestHandleReceivedData_SplitStartSequenceWithJunk(t *testing.T) {
	tc := NewTransportController()
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55}
	// First call: only 0x55
	got := tc.HandleReceivedData([]byte{0x00, 0x00, 0x12, 0x70, 0x55})

	assert.Nil(t, got, "expected nil packet list")
	// Second call: 0xaa and 0x02, should complete start and fill
	got = tc.HandleReceivedData([]byte{0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55})

	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 1, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want, "expected packet bytes %v, got %v", want, got[0].ToBytes())
}

func TestHandleReceivedData_MultiplePackets(t *testing.T) {
	tc := NewTransportController()
	// Two packets in one buffer
	data := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x05, 0xA7, 0xD5, 0xaa, 0x55, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x06, 0x95, 0x4E, 0xaa, 0x55}
	want := [][]byte{
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55},
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x05, 0xA7, 0xD5, 0xaa, 0x55},
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x06, 0x95, 0x4E, 0xaa, 0x55},
	}
	got := tc.HandleReceivedData(data)
	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 3, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want[0], "expected packet bytes %v, got %v", want[0], got[0].ToBytes())
	assert.Equal(t, got[1].ToBytes(), want[1], "expected packet bytes %v, got %v", want[1], got[1].ToBytes())
	assert.Equal(t, got[2].ToBytes(), want[2], "expected packet bytes %v, got %v", want[2], got[2].ToBytes())
}

func TestHandleReceivedData_MultiplePacketsWithJunk(t *testing.T) {
	tc := NewTransportController()
	// Two packets in one buffer
	data := []byte{0x00, 0x01, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55,
		0x00, 0x34, 0x45, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x05, 0xA7, 0xD5, 0xaa, 0x55, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x06, 0x95, 0x4E, 0xaa, 0x55, 0x00, 0x03}
	want := [][]byte{
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55},
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x05, 0xA7, 0xD5, 0xaa, 0x55},
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x06, 0x95, 0x4E, 0xaa, 0x55},
	}
	got := tc.HandleReceivedData(data)
	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 3, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want[0], "expected packet bytes %v, got %v", want[0], got[0].ToBytes())
	assert.Equal(t, got[1].ToBytes(), want[1], "expected packet bytes %v, got %v", want[1], got[1].ToBytes())
	assert.Equal(t, got[2].ToBytes(), want[2], "expected packet bytes %v, got %v", want[2], got[2].ToBytes())
}

func TestHandleReceivedData_NoStartSequence(t *testing.T) {
	tc := NewTransportController()
	data := []byte{0x01, 0x02, 0x03}
	got := tc.HandleReceivedData(data)
	assert.Nil(t, got, "expected nil packet list for data without start sequence")

}

func TestHandleReceivedData_PartialPacket(t *testing.T) {
	tc := NewTransportController()
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55}
	// Only part of the packet arrives

	got := tc.HandleReceivedData([]byte{0x55, 0xaa})
	assert.Nil(t, got, "expected nil packet list for partial packet")
	// Second call: more data arrives, but still not complete
	got = tc.HandleReceivedData([]byte{0x02, 0x00, 0x03, 0x04, 0xB6})
	assert.Nil(t, got, "expected nil packet list for partial packet")

	// Third call: more data arrives
	got = tc.HandleReceivedData([]byte{0x5C, 0xaa})
	assert.Nil(t, got, "expected nil packet list for partial packet")

	// Fourth call: complete packet arrives
	got = tc.HandleReceivedData([]byte{0x55})
	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 1, "expected 1 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want, "expected packet bytes %v, got %v", want, got[0].ToBytes())
}

func TestHandleReceivedData_PartialPacketMultiplePacket(t *testing.T) {
	tc := NewTransportController()
	want := [][]byte{
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x04, 0xB6, 0x5C, 0xaa, 0x55},
		{0x55, 0xaa, 0x02, 0x00, 0x03, 0x05, 0xA7, 0xD5, 0xaa, 0x55},
	}
	// Only part of the packet arrives

	got := tc.HandleReceivedData([]byte{0x55, 0xaa})
	assert.Nil(t, got, "expected nil packet list for partial packet")
	// Second call: more data arrives, but still not complete
	got = tc.HandleReceivedData([]byte{0x02, 0x00, 0x03, 0x04, 0xB6})
	assert.Nil(t, got, "expected nil packet list for partial packet")

	// Third call: more data arrives
	got = tc.HandleReceivedData([]byte{0x5C, 0xaa})
	assert.Nil(t, got, "expected nil packet list for partial packet")

	// Fourth call: complete packet arrives
	got = tc.HandleReceivedData([]byte{0x55, 0x55, 0xaa, 0x02, 0x00, 0x03, 0x05, 0xA7, 0xD5, 0xaa, 0x55})
	assert.NotEmpty(t, got, "expected non-nil packet list")
	assert.Equal(t, len(got), 2, "expected 2 packet, got %d", len(got))
	assert.Equal(t, got[0].ToBytes(), want[0], "expected packet bytes %v, got %v", want[0], got[0].ToBytes())
	assert.Equal(t, got[1].ToBytes(), want[1], "expected packet bytes %v, got %v", want[1], got[1].ToBytes())
}
