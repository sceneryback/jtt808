package location

import (
	"bytes"
	"fmt"

	"github.com/sceneryback/jtt808/utils"
)

type Wifi struct {
	MacAddress     string `json:"MAC 地址,omitempty"`
	SignalStrength uint8  `json:"信号强度,omitempty"`
}

type AdditionalInfoWifis struct {
	Wifis []*Wifi `json:"WiFi 列表,omitempty"`

	data []byte
}

func (a *AdditionalInfoWifis) Id() uint8 {
	return uint8(0x54)
}

func (a *AdditionalInfoWifis) Length() uint8 {
	return uint8(len(a.Wifis)*7 + 1)
}

func (a *AdditionalInfoWifis) Info() []byte {
	return a.data
}

func (a *AdditionalInfoWifis) Human() string {
	var buf bytes.Buffer

	buf.WriteString("WiFi list:\n")

	for i := range a.Wifis {
		buf.WriteString(fmt.Sprintf("%s %d\n", a.Wifis[i].MacAddress, a.Wifis[i].SignalStrength))
	}

	return buf.String()
}

type wifiCodec struct {
}

func (w *wifiCodec) Decode(data []byte) (*AdditionalInfoWifis, error) {
	var wifis = AdditionalInfoWifis{
		data: data,
	}

	wifisNum := int(data[0])
	data = data[1:]

	for i := 0; i < wifisNum; i++ {
		wifis.Wifis = append(wifis.Wifis, &Wifi{
			MacAddress:     utils.HexBytesToMacAddr(data[i*7 : i*7+6]),
			SignalStrength: ^uint8(data[i*7+6]) + 1,
		})
	}

	return &wifis, nil
}
