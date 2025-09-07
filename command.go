package driver100

import "encoding/binary"

const (
	headerByte0    = 0xF8
	headerByte1    = 0x55
	headerByte2    = 0xCE
	lenByte3       = 0x00
	lenByte4       = 0x01
	cmdGetMassa    = 0x23
	cmdGetScalePar = 0x75
)

// CMD_GET_MASSA - команда получения массы
// Вначале я использовал append для построения сообщения, но для быстроты работы, переписал на запись по индексам
func buildGetMassa() []byte {
	// Команда состоит из 8 байт, лимитируем емкость среза
	msg := make([]byte, 8)
	//обязательные байты заголовка
	//msg = append(msg, 0xF8, 0x55, 0xCE) - так было изначально
	msg[0] = headerByte0
	msg[1] = headerByte1
	msg[2] = headerByte2
	//длину сообщения записываем отдельными байтами, иначе при записи 0x0001 будет записан лишь один байт
	msg[3] = lenByte3
	msg[4] = lenByte4
	//код CMD_GET_MASSA
	msg[5] = cmdGetMassa
	//вычисляем контрольную сумму и добавляем в конец сообщения
	crc := crc16alg(msg[5:6])
	//В спецификации не указано какой порядок байт используется(little или bid endian).
	//1-согласно Гугл, АРМ процессоры используют little endian
	//2- в примере протокола int16 Len вначале упоминается младший байт, что характерно для little endian
	binary.LittleEndian.PutUint16(msg[:6], crc)
	return msg
}

// CMD_GET_SCALE_PAR - команда получения параметров весового устройства
func buildGetScalePar() []byte {
	msg := make([]byte, 8)
	msg[0] = headerByte0
	msg[1] = headerByte1
	msg[2] = headerByte2
	msg[3] = lenByte3
	msg[4] = lenByte4
	//код CMD_GET_SCALE_PAR
	msg[5] = cmdGetScalePar
	crc := crc16alg(msg[5:6])
	binary.LittleEndian.PutUint16(msg[:6], crc)
	return msg
}
