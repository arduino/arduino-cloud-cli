package serial

var (
	// msgStart is the initial byte sequence of every packet
	msgStart = [2]byte{0x55, 0xAA}
	// msgEnd is the final byte sequence of every packet
	msgEnd = [2]byte{0xAA, 0x55}
)

const (
	// Position of payload field
	payloadField = 5
	// Position of payload length field
	payloadLenField = 3
	// Length of payload length field
	payloadLenFieldLen = 2
	// Length of the signature field
	crcFieldLen = 2
)

// MsgType indicates the type of the packet
type MsgType byte

const (
	None MsgType = iota
	Cmd
	Data
	Response
)

// Command indicates the command that should be
// executed on the board to be provisioned.
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
