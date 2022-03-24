package response

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/sceneryback/jtt808/codec/protocol"
	"github.com/sceneryback/jtt808/log"
)

var (
	ErrBodyNotClientResponse = errors.New("body is not client response")
	ErrBodyNotServerResponse = errors.New("body is not server response")
)

var logger = log.Logger

type clientResponseCodec struct {
}

func NewClientResponseCodec() *clientResponseCodec {
	return &clientResponseCodec{}
}

// 0x8001
func (c *clientResponseCodec) Encode(b protocol.Body) ([]byte, error) {
	r, ok := b.(*ClientResponse)
	if !ok {
		logger.Error("message body is not client response")
		return nil, ErrBodyNotClientResponse
	}

	var res bytes.Buffer
	binary.Write(&res, binary.BigEndian, r.SerialNum)
	binary.Write(&res, binary.BigEndian, r.ID)
	res.Write([]byte{r.Result})
	return res.Bytes(), nil
}

func (c *clientResponseCodec) Decode(data []byte) (protocol.Body, error) {
	return nil, nil
}

type serverResponseCodec struct {
}

func NewServerResponseCodec() *serverResponseCodec {
	return &serverResponseCodec{}
}

// 0x8001
func (c *serverResponseCodec) Encode(b protocol.Body) ([]byte, error) {
	r, ok := b.(*ServerResponse)
	if !ok {
		logger.Error("message body is not server response")
		return nil, ErrBodyNotServerResponse
	}

	var res bytes.Buffer
	binary.Write(&res, binary.BigEndian, r.SerialNum)
	binary.Write(&res, binary.BigEndian, r.ID)
	res.Write([]byte{r.Result})
	return res.Bytes(), nil
}

func (c *serverResponseCodec) Decode(data []byte) (protocol.Body, error) {
	return nil, nil
}
