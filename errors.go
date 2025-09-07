package driver100

// мапа для вывода ошибок по коду
var ErrorCodes = map[byte]string{
	0x07: "Команда не поддерживается",
	0x08: "Нагрузка на весовом устройстве превышает НПВ",
	0x09: "Весовое устройство не в режиме взвешивания",
	0x0A: "Ошибка входных данных",
	0x0B: "Ошибка сохранения данных",
	0x10: "Интерфейс WiFi не поддерживается",
	0x11: "Интерфейс Ethernet не поддерживается",
	0x15: "Установка >0< невозможна",
	0x17: "Нет связи с модулем взвешивающим",
	0x18: "Установлена нагрузка на платформу при включении весового устройства",
	0x19: "Весовое устройство неисправно",
}

// использование ошибок как констант в коде идиоматично
const (
	ErrCmdNotSupported         = 0x07
	ErrOverload                = 0x08
	ErrNotWeighingMode         = 0x09
	ErrInputData               = 0x0A
	ErrSaveData                = 0x0B
	ErrWiFiNotSupported        = 0x10
	ErrEthernetNotSupported    = 0x11
	ErrSetZeroImpossible       = 0x15
	ErrNoConnectionToModule    = 0x17
	ErrLoadOnPlatformAtStartup = 0x18
	ErrDeviceFault             = 0x19
	ErrRunCommand              = 0x28
	ErrUnknown                 = 0xF0
)
