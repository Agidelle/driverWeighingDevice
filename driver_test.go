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

func TestParsingWeight(t *testing.T) {
	// пример буфера: 4 байта веса, 1 деление, 1 стабильность, 1 нетто, 1 ноль, 4 байта тары
	buf := []byte{
		0x15, 0xCD, 0x5B, 0x07, // вес: 123456789
		0x02,                   // деление
		0x01,                   // стабильность
		0x01,                   // нетто
		0x00,                   // ноль
		0x04, 0x03, 0x02, 0x01, // тара: 0xDDCCBBAA
	}
	length := 12

	input, err := parsingWeight(buf, length)
	if err != nil {
		t.Fatalf("Ошибка парсинга: %v", err)
	}

	if input.Weight != 123456789 {
		t.Errorf("Ожидался вес 123456789, получено %d", input.Weight)
	}
	if input.Division != 0x02 {
		t.Errorf("Ожидалось деление 0x02, получено 0x%X", input.Division)
	}
	if input.Stable != 0x01 {
		t.Errorf("Ожидалась стабильность 0x01, получено 0x%X", input.Stable)
	}
	if input.Net != 0x01 {
		t.Errorf("Ожидалось нетто 0x01, получено 0x%X", input.Net)
	}
	if input.Zero != 0x00 {
		t.Errorf("Ожидался ноль 0x00, получено 0x%X", input.Zero)
	}
	if input.Tare != int32(0x01020304) {
		t.Errorf("Ожидалась тара 0xDDCCBBAA, получено 0x%X", input.Tare)
	}
}

func TestParsingScalePar(t *testing.T) {
	// 8 параметров, каждый заканчивается 0x0D 0x0A
	buf := append(
		append([]byte("PMaxValue\x0D\x0A"), []byte("PMinValue\x0D\x0A")...),
		append(
			append([]byte("PEValue\x0D\x0A"), []byte("PTValue\x0D\x0A")...),
			append(
				append([]byte("FixValue\x0D\x0A"), []byte("CalcodeValue\x0D\x0A")...),
				append([]byte("PoVerValue\x0D\x0A"), []byte("PoSummValue\x0D\x0A")...)...,
			)...,
		)...,
	)
	length := len(buf)

	input, err := parsingScalePar(buf, length)
	if err != nil {
		t.Fatalf("Ошибка парсинга: %v", err)
	}

	if input.PMax != "PMaxValue" {
		t.Errorf("Ожидалось PMaxValue, получено %s", input.PMax)
	}
	if input.PMin != "PMinValue" {
		t.Errorf("Ожидалось PMinValue, получено %s", input.PMin)
	}
	if input.PE != "PEValue" {
		t.Errorf("Ожидалось PEValue, получено %s", input.PE)
	}
	if input.PT != "PTValue" {
		t.Errorf("Ожидалось PTValue, получено %s", input.PT)
	}
	if input.Fix != "FixValue" {
		t.Errorf("Ожидалось FixValue, получено %s", input.Fix)
	}
	if input.Calcode != "CalcodeValue" {
		t.Errorf("Ожидалось CalcodeValue, получено %s", input.Calcode)
	}
	if input.PoVer != "PoVerValue" {
		t.Errorf("Ожидалось PoVerValue, получено %s", input.PoVer)
	}
	if input.PoSumm != "PoSummValue" {
		t.Errorf("Ожидалось PoSummValue, получено %s", input.PoSumm)
	}
}
