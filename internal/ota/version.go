package ota

// Version contains all the OTA header information
// Check out https://arduino.atlassian.net/wiki/spaces/RFC/pages/1616871540/OTA+header+structure for more
// information on the OTA header specs.
type Version struct {
	HeaderVersion   uint8
	Compression     bool
	Signature       bool
	Spare           uint8
	PayloadTarget   uint8
	PayloadMayor    uint8
	PayloadMinor    uint8
	PayloadPatch    uint8
	PayloadBuildNum uint32
}

// AsBytes builds a 8 byte length representation of the Version Struct for the OTA update.
func (v *Version) AsBytes() []byte {
	version := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	// Set compression
	if v.Compression {
		version[7] = 0x40
	}

	// Other field are currently not implemented ¯\_(ツ)_/¯

	return version
}
