package location

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type UnknownInfo struct {
	id     uint8
	length uint8
	body   []byte
}

func (u *UnknownInfo) Id() uint8 {
	return u.id
}

func (u *UnknownInfo) Length() uint8 {
	return u.length
}

func (u *UnknownInfo) Info() []byte {
	return u.body
}

func (u *UnknownInfo) Human() string {
	var buf bytes.Buffer

	buf.WriteString("Unknown info:\n")

	buf.WriteString(fmt.Sprintf("id: %d\n", u.id))
	buf.WriteString(fmt.Sprintf("length: %d\n", u.length))
	buf.WriteString(fmt.Sprintf("body: %s\n", hex.EncodeToString(u.body)))

	return buf.String()
}

type unknownInfoCodec struct {
}

func (c *unknownInfoCodec) Decode(data []byte) (*UnknownInfo, error) {
	var info UnknownInfo

	info.id = uint8(data[0])
	info.length = uint8(data[1])
	info.body = data[2:]

	return &info, nil
}
