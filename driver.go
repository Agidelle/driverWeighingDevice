package driver100

import (
	"encoding/binary"
	"fmt"
	"github.com/sigurn/crc16"
	"net"
	"time"
)

const (
	// Таймауты для операций с весами, чтобы не было магических чисел в коде
	timeoutConn  = 1 * time.Second
	timeoutRead  = 2 * time.Second
	timeoutWrite = 2 * time.Second
	//коды успешного выполнения команд
	successfulMassa = 0x24
	successfulScale = 0x76
)

// Driver Интерфейс для инкапсуляции драйвера
type Driver interface {
	OpenConnection(ip, port string) error
	CloseConnection() error
	ReadWeight() (*MassaInput, error)
	ReadScaleParameters() (*ScaleParameters, error)
}

// NewDriver Конструктор драйвера, используем интерфейс
func NewDriver() Driver {
	return &driver{}
}

// OpenConnection - открытие TCP соединения с весовым устройством
// предположил что Ip и port будут передаваться в функцию, т.к. есть команда запроса ethernet параметров,
// а их настройка осуществляется через usb/rs-232.
func (d *driver) OpenConnection(ip, port string) error {
	//собираю строку подключения
	connStr := fmt.Sprintf("%s:%s", ip, port)
	//открываю соединение с таймаутом
	conn, err := net.DialTimeout("tcp", connStr, timeoutConn)
	if err != nil {
		return err
	}
	defer conn.Close()
	//сохраняю соединение в драйвере
	d.conn = conn
	return nil
}

// CloseConnection - закрытие TCP соединения с весовым устройством
func (d *driver) CloseConnection() error {
	//проверка, что соединение открыто
	if d.conn != nil {
		return d.conn.Close()
	}
	return fmt.Errorf("соединения не существует")
}

// readWrite - отправка команды и чтение ответа
// универсальная функция для отправки команд и чтения ответов
// не экспортируется, нужна для внутреннего использования
// возвращает буфер с ответом, количество прочитанных байт и ошибку
func (d *driver) readWrite(cmd []byte) ([]byte, int, error) {
	//проверка, что соединение открыто
	if d.conn == nil {
		return nil, 0, fmt.Errorf("соединения не существует")
	}
	//установка таймаута на запись
	err := d.conn.SetWriteDeadline(time.Now().Add(timeoutWrite))
	if err != nil {
		return nil, 0, err
	}
	//отправка команды, ф-ция buildGetMassa собирает строку команды CMD_GET_MASSA
	wByte, err := d.conn.Write(cmd)
	if err != nil {
		return nil, 0, err
	}
	//проверка, что отправлено 8 байт
	if wByte != 8 {
		return nil, 0, fmt.Errorf("неправильное кол-во отправленных байт: %d, вместо 8", wByte)
	}
	//установка таймаута на чтение
	err = d.conn.SetReadDeadline(time.Now().Add(timeoutRead))
	if err != nil {
		return nil, 0, err
	}
	//чтение ответа, согласно спецификации ответ содержит 8-20 байт
	//использование ReadFull не подходит, т.к. неизвестно сколько байт будет в ответе
	buf := make([]byte, 20)
	lastByte, err := d.conn.Read(buf)
	if err != nil {
		return nil, 0, err
	}
	//проверки ответа, повторяющийся код вынесен в отдельную функцию
	err = checkResponse(buf, lastByte)
	if err != nil {
		return nil, 0, err
	}
	return buf, lastByte, nil
}

// ReadWeight - получение текущих параметров устройства
func (d *driver) ReadWeight() (*MassaInput, error) {
	//проверка, что соединение открыто
	buf, lastByte, err := d.readWrite(buildGetMassa())
	if err != nil {
		return nil, err
	}
	//парсинг ответа
	switch buf[5] {
	case successfulMassa:
		//передаю срез и длину без заголовка, длины, кода команды и CRC
		input, err := parsingWeight(buf[6:lastByte-2], lastByte-8)
		if err != nil {
			return nil, err
		}
		return input, nil
	case ErrRunCommand:
		//парсинг ошибки выполнения команды
		switch buf[6] {
		case ErrOverload, ErrNotWeighingMode, ErrNoConnectionToModule, ErrLoadOnPlatformAtStartup, ErrDeviceFault:
			// обработка известных ошибок
			return nil, fmt.Errorf("ошибка выполнения команды %X: %X:%s", ErrRunCommand, buf[6], ErrorCodes[buf[6]])
		default:
			return nil, fmt.Errorf("непредусмотренный код ошибки: 0x%X", buf[6])
		}
	case ErrUnknown:
		//парсинг неизвестной команды CMD_NACK: 0xF0
		return nil, fmt.Errorf("неизвестная ошибка: %X:%s", buf[6], ErrorCodes[buf[6]])
	default:
		return nil, fmt.Errorf("неизвестный код ответа: 0x%X", buf[5])
	}
}

// parsingWeight - парсинг ответа от весового устройства
// был вариант написать общую функцию парсинга. Но для читаемости и простоты поддержки решил сделать отдельную функцию для каждого типа ответа
func parsingWeight(buf []byte, length int) (input *MassaInput, err error) {
	if buf == nil {
		return nil, fmt.Errorf("пустой буфер")
	}
	if len(buf) != length {
		return nil, fmt.Errorf("неправильная длина буфера: %d, вместо %d", len(buf), length)
	}
	input = &MassaInput{}
	weight := binary.LittleEndian.Uint32(buf[0:4])
	input.Weight = int32(weight)

	// в примечаниях к CMD_ACK_MASSA указано, что поле Tare может отсутствовать.
	// в связи с этим делаем проверку длины буфера
	// то что в поле Len указано 0х0009, я так понимаю, это лишь 1 из вариантов ответа(мин длина, от Command до CRC, аналогично CMD_ACK_SCALE_PAR)
	if length == 12 {
		tare := binary.LittleEndian.Uint32(buf[8:])
		input.Tare = int32(tare)
	}

	input.Division = buf[4]
	input.Stable = buf[5]
	input.Net = buf[6]
	input.Zero = buf[7]
	return input, nil
}

// ReadScaleParameters - получение базовых параметров устройства
func (d *driver) ReadScaleParameters() (*ScaleParameters, error) {
	buf, lastByte, err := d.readWrite(buildGetScalePar())
	if err != nil {
		return nil, err
	}
	//парсинг ответа
	// я принимаю как аксиому, то что, если проверка CRC прошла успешно, то дополнительные проверки на выход за границы во время парсинга не нужны
	// и данные целые и корректные
	switch buf[5] {
	case successfulScale:
		//передаю срез и длину без заголовка, длины, кода команды и CRC
		input, err := parsingScalePar(buf[6:lastByte-2], lastByte-8)
		if err != nil {
			return nil, err
		}
		return input, nil
	case ErrRunCommand:
		//парсинг ошибки выполнения команды
		switch buf[6] {
		case ErrNoConnectionToModule:
			return nil, fmt.Errorf("ошибка выполнения команды %X: %X:%s", ErrRunCommand, buf[6], ErrorCodes[buf[6]])
		default:
			return nil, fmt.Errorf("непредусмотренный код ошибки: 0x%X", buf[6])
		}
	case ErrUnknown:
		//парсинг неизвестной ошибки CMD_NACK: 0xF0
		return nil, fmt.Errorf("неизвестная ошибка: %X:%s", buf[6], ErrorCodes[buf[6]])
	default:
		return nil, fmt.Errorf("неизвестный код ответа: 0x%X", buf[5])
	}
}

func parsingScalePar(buf []byte, length int) (input *ScaleParameters, err error) {
	// проверка на пустой буфер и длину
	if buf == nil {
		return nil, fmt.Errorf("пустой буфер")
	}
	if len(buf) != length {
		return nil, fmt.Errorf("неправильная длина буфера: %d, вместо %d", len(buf), length)
	}
	input = &ScaleParameters{}
	slicePar := make([][]byte, 8)
	offset := 0
	for i := 0; i < 8; i++ {
		//делаю срез от оффсета до конца, т.к. в ф-цию уже передан срез без CRC
		par, l := readStop(buf[offset:])
		slicePar[i] = par
		offset += l
	}

	input.PMax = string(slicePar[0])
	input.PMin = string(slicePar[1])
	input.PE = string(slicePar[2])
	input.PT = string(slicePar[3])
	input.Fix = string(slicePar[4])
	input.Calcode = string(slicePar[5])
	input.PoVer = string(slicePar[6])
	input.PoSumm = string(slicePar[7])
	return input, nil
}

func readStop(buf []byte) ([]byte, int) {
	text := make([]byte, 0, 20)
	count := 0
	//проверка на пустой буфер
	if buf == nil {
		return text, 0
	}
	//предварительное выделение памяти, чтобы избежать частых аллокаций
	//максимальная длина параметра в спецификации 20 байт
	//да есть параметры меньшей длины, можно было бы сделать для каждого параметра свою функцию чтения,
	//либо передавать константу длины конкретного параметра.
	//Но т.к. у всех параметров есть разделитель 0x0D 0x0A, решил сделать универсальную функцию.
	//Лишних циклов не будет, т.к. цикл прервется при нахождении разделителя
	for i := 0; i < 19; i++ {
		if i+2 <= len(buf) && buf[i] == 0x0D && buf[i+1] == 0x0A {
			count = i + 2
			break
		}
		text = append(text, buf[i])
	}
	return text, count
}

func checkResponse(buf []byte, lastByte int) error {
	//headers должны быть 0xF8 0x55 0xCE
	if buf[0] != headerByte0 || buf[1] != headerByte1 || buf[2] != headerByte2 {
		return fmt.Errorf("неправильный заголовок ответа: 0x%X 0x%X 0x%X", buf[0], buf[1], buf[2])
	}
	//проверка на корректную длину ответа согласно спецификации протокола
	switch buf[5] {
	case successfulMassa:
		if lastByte < 8 || lastByte > 20 {
			return fmt.Errorf("неправильное кол-во принятых байт: %d", lastByte)
		}
	case successfulScale:
		if lastByte < 8 || lastByte > 104 {
			return fmt.Errorf("неправильное кол-во принятых байт: %d", lastByte)
		}
	}

	//проверка на совпадение длины ответа и байтов длины
	lengthResponse := int(binary.LittleEndian.Uint16(buf[3:5]))
	if lengthResponse != lastByte-7 {
		return fmt.Errorf("неправильная длина ответа: %d, вместо %d", lastByte-7, lengthResponse)
	}
	//проверка CRC
	if !checkCRC16CCITT(buf[5:lastByte-2], buf[lastByte-2:lastByte]) {
		return fmt.Errorf("неправильная контрольная сумма")
	}
	return nil
}

// сheckCRC16CCITT проверка CRC, решил использовать библиотеку-обёртку над hash, т.к. алгоритм популярный
// юнит тест на проверку алгоритма реализован, тестовые данные из примера, с проверкой на CRC True, а не распространенный CRC Bad.
// В протоколе указана ссылка, на обсуждение проблем алгоритма.
// после долгого изучения форума, пришел к выводу, что использование CRC16_CCITT_FALSE неправильно.
// решением оказался протокол реализация от фуджицу CRC-16/SPI-FUJITSU
// важное примечание к начальному значению: "Init value is equivalent to an augment of 0xFFFF prepended to the message."
func checkCRC16CCITT(data []byte, crc []byte) bool {
	check := crc16alg(data)
	return check == binary.LittleEndian.Uint16(crc)
}

func crc16alg(data []byte) uint16 {
	table := crc16.MakeTable(crc16.CRC16_AUG_CCITT)
	check := crc16.Checksum(data, table)
	return check
}
