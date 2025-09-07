package driver100

import (
	"testing"
)

// проверка на True CRC16-CCITT из https://srecord.sourceforge.net/crc16-ccitt.html#long-hand
func TestCrc16alg(t *testing.T) {
	data := []byte("123456789")
	expectedCRC := uint16(0xE5CC)

	actualCRC := crc16alg(data)
	if actualCRC != expectedCRC {
		t.Errorf("Ожидалось CRC16CCITT: 0x%X, получено: 0x%X", expectedCRC, actualCRC)
	}
}
