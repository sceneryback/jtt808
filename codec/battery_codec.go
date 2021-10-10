package codec

import (
	"bytes"
	"fmt"
)

type Battery struct {
	Percentage uint8
	Extention  uint8
	Raw        []byte
}

func (b *Battery) Id() uint8 {
	return uint8(0x56)
}

func (b *Battery) Length() uint8 {
	return 3
}

func (b *Battery) Info() []byte {
	return b.Raw
}

func (b *Battery) Human() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("battery: %.2f%%\n", float64(b.Percentage)*100/10))

	return buf.String()
}

type batteryCodec struct {
}

func (w *batteryCodec) Decode(data []byte) (*Battery, error) {
	return &Battery{
		Percentage: uint8(data[0]),
		Extention:  uint8(data[1]),
		Raw:        data,
	}, nil
}
