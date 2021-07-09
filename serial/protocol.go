package serial

var (
	msgStart = [2]byte{0x55, 0xAA}
	msgEnd   = [2]byte{0xAA, 0x55}
)

type MsgType byte

const (
	None MsgType = iota
	Cmd
	Data
	Response
)

type Command byte

const (
	SketchInfo Command = iota + 1
	CSR
	Locked
	GetLocked
	WriteCrypto
	BeginStorage
	SetDeviceID
	SetYear
	SetMonth
	SetDay
	SetHour
	SetValidity
	SetCertSerial
	SetAuthKey
	SetSignature
	EndStorage
	ReconstructCert
)
