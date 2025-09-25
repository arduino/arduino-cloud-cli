package iotapiraw

type BoardType struct {
	FQBN                 *string  `json:"fqbn,omitempty"`
	Label                string   `json:"label"`
	MinProvSketchVersion *string  `json:"min_provisioning_sketch_version,omitempty"`
	MinWiFiVersion       *string  `json:"min_provisioning_wifi_version,omitempty"`
	Provisioning         *string  `json:"provisioning,omitempty"`
	Tags                 []string `json:"tags"`
	Type                 string   `json:"type"`
	Vendor               string   `json:"vendor"`
	OTAAvailable         *bool    `json:"ota_available,omitempty"`
}

type BoardTypeList []BoardType

type Prov2SketchBinRes struct {
	Binary   string `json:"bin"`
	FileName string `json:"filename"`
	FQBN     string `json:"fqbn"`
	Name     string `json:"name"`
	SHA256   string `json:"sha256"`
}
