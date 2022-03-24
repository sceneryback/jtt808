package location

type BasicInfo struct {
	Alert     uint32  `json:"报警标志,omitempty"`
	State     uint32  `json:"状态,omitempty"`
	Latitude  float32 `json:"纬度,omitempty"`
	Longitude float32 `json:"经度,omitempty"`
	Altitude  uint16  `json:"高程,omitempty"`
	Speed     uint16  `json:"速度,omitempty"`
	Direction uint16  `json:"方向,omitempty"`
	Timestamp string  `json:"时间,omitempty"`
}

type LocationAdditionalInfo interface {
	Id() uint8
	Length() uint8
	Info() []byte
}

type LocationMsgBody struct {
	Basic              *BasicInfo               `json:"位置基本信息,omitempty"`
	BasicHex           string                   `json:"位置基本信息(十六进制),omitempty"`
	AdditionalInfos    []LocationAdditionalInfo `json:"位置附加信息项列表,omitempty"`
	AdditionalInfosHex string                   `json:"位置附加信息项列表(十六进制),omitempty"`

	data []byte `json:"-"`
}

func (l LocationMsgBody) Data() []byte {
	return l.data
}
