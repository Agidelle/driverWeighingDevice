package driver100

import "net"

type driver struct {
	conn net.Conn
}

// ScaleParameters - параметры весового устройства.
// Был вопрос использовать ли экспортируемую структуру с полями. Данная структура нужна для хранения базовых параметров и информации весов
// и не используется для логики драйвера согласно спецификации. Напрашиваются проверки по нагрузке весов/тары, но это должно выполняться на уровне бизнес-логики сервиса, а не драйвера,
// поэтому я не вижу смысла инкапсулировать поля в неэкспортируемую структуру, что привело бы к необходимости писать Get функции для полей.
// Хранение в структуре драйвера "driver" в связи с этим не имеет смысла.
type ScaleParameters struct {
	PMax    string
	PMin    string
	PE      string
	PT      string
	Fix     string
	Calcode string
	PoVer   string
	PoSumm  string
}

// MassaInput - данные полученные из CMD_GET_MASSA
type MassaInput struct {
	Weight   int32
	Division byte
	Stable   byte
	Net      byte
	Zero     byte
	Tare     int32
}
