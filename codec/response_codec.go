package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ErrBodyNotResponse = errors.New("body is not server response")
)

type responseCodec struct {
}

// 0x8001
func (c *responseCodec) Encode(b Body) ([]byte, error) {
	r, ok := b.(*ServerResponse)
	if !ok {
		return nil, ErrBodyNotResponse
	}

	var res bytes.Buffer
	binary.Write(&res, binary.BigEndian, r.SerialNum)
	binary.Write(&res, binary.BigEndian, r.ID)
	res.Write([]byte{r.Result})
	return res.Bytes(), nil
}

func (c *responseCodec) Decode(data []byte) (Body, error) {
	return nil, nil
}
