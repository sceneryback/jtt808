package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ErrChecksumFailed        = errors.New("failed to verify checksum")
	ErrDecodeHeaderFailed    = errors.New("failed to decode header")
	ErrMessageIdNotSupported = errors.New("message id not supported yet")
)

type Codec interface {
	Encode(*Message) ([]byte, error)
	Decode([]byte) (*Message, error)
}

type HeaderCodec interface {
	Encode(header *Header) ([]byte, error)
	Decode([]byte) (*Header, error)
}

type BodyCodec interface {
	Encode(Body) ([]byte, error)
	Decode([]byte) (Body, error)
}

type CodecConfig struct {
	Version string
}

type codec struct {
	header HeaderCodec
	body   BodyCodec
}

func NewCodec(cfg *CodecConfig) (Codec, error) {
	return &codec{
		header: &headerCodec{},
	}, nil
}

func (c *codec) Encode(msg *Message) ([]byte, error) {
	var buf bytes.Buffer

	headerBytes, err := c.header.Encode(msg.H)
	if err != nil {
		return nil, err
	}

	buf.Write(headerBytes)

	switch msg.H.MessageId {
	case 0x8001:
		c.body = &responseCodec{}
	default:
		return nil, ErrMessageIdNotSupported
	}

	bodyBytes, err := c.body.Encode(msg.B)
	if err != nil {
		return nil, err
	}

	buf.Write(bodyBytes)

	headerBodyBytes := buf.Bytes()

	buf.WriteByte(c.checksum(headerBodyBytes))

	var res = []byte{0x7e}
	res = append(res, c.escape(buf.Bytes())...)
	res = append(res, 0x7e)

	return res, nil
}

/*
rules：
0x7e => 0x7d02
0x7d => 0x7d01
*/
func (*codec) escape(data []byte) []byte {
	var result []byte
	var l = len(data)
	// escape
	for i := 0; i < l; i++ {
		if data[i] == 0x7e {
			result = append(result, 0x7d, 0x02)
		} else if data[i] == 0x7d {
			result = append(result, 0x7d, 0x01)
		} else {
			result = append(result, data[i])
		}
	}
	return result
}

func (*codec) unescape(data []byte) []byte {
	var result []byte
	var l = len(data)
	// unescape, i.e. restore
	for i := 0; i < l; i++ {
		if data[i] == 0x7d && data[i+1] == 0x2 {
			result = append(result, 0x7e)
			i++
		} else if data[i] == 0x7d && data[i+1] == 0x1 {
			result = append(result, 0x7d)
			i++
		} else {
			result = append(result, data[i])
		}
	}
	return result
}

// headerBodyBytes Contains original unescaped header and body bytes
func (c *codec) checksum(headerBodyBytes []byte) byte {
	var realsum = headerBodyBytes[0]
	var msgLength = len(headerBodyBytes)
	for i := 1; i < msgLength; i++ {
		realsum = realsum ^ headerBodyBytes[i]
	}
	return realsum
}

// unescaped(header+body+checksum)
func (c *codec) checksumVerified(unescapedMsg []byte) bool {
	unescapedLength := len(unescapedMsg)

	var checksum = unescapedMsg[unescapedLength-1:][0]
	realsum := c.checksum(unescapedMsg[0 : unescapedLength-1])

	return realsum == checksum
}

func (c *codec) decodeHeader(h []byte) (*Header, error) {
	var header Header

	var msgIdBytes = h[:2]
	var msgAttrBytes = h[2:4]
	var phoneBytes = h[4:10]
	var serialNumBytes = h[10:12]

	err := binary.Read(bytes.NewReader(msgIdBytes), binary.BigEndian, &header.MessageId)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(phoneBytes), binary.BigEndian, &header.Phone)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(serialNumBytes), binary.BigEndian, &header.SerialNum)
	if err != nil {
		return nil, err
	}

	header.Attr = &BodyAttr{}
	header.Attr.Preserved = uint8(msgAttrBytes[0] >> 6)
	if msgAttrBytes[0]&0x20 == 1 {
		header.Attr.SegmentationEnabled = true
	}
	var encrypt = msgAttrBytes[0] & 0x1C
	if encrypt == 1 {
		header.Attr.EncryptionMethod = "RSA"
	}
	header.Attr.BodyLength = uint16((msgAttrBytes[0]&0x03)<<8 + uint8(msgAttrBytes[1]))

	if header.Attr.SegmentationEnabled {
		var segmentBytes = h[12:]

		header.SegInfo = &SegmentInfo{}

		err = binary.Read(bytes.NewReader(segmentBytes[:2]), binary.BigEndian, &header.SegInfo.TotalSegments)
		if err != nil {
			return nil, err
		}
		err = binary.Read(bytes.NewReader(segmentBytes[2:]), binary.BigEndian, &header.SegInfo.SegmentNum)
		if err != nil {
			return nil, err
		}
	}

	return &header, nil
}

func (c *codec) trimIdentifiers(data []byte) []byte {
	if data[0] == 0x7e {
		data = data[1:]
	}
	if data[len(data)-1] == 0x7e {
		data = data[:len(data)-1]
	}
	return data
}

func (c *codec) Decode(data []byte) (*Message, error) {
	data = c.trimIdentifiers(data)

	unescapedData := c.unescape(data)

	if !c.checksumVerified(unescapedData) {
		return nil, ErrChecksumFailed
	}

	var msg Message

	header, err := c.header.Decode(unescapedData[:MessageHeaderMaxLength])
	if err != nil {
		return nil, ErrDecodeHeaderFailed
	}
	msg.H = header

	switch msg.H.MessageId {
	case 0x0200:
		c.body = &locationCodec{}
	default:
		return nil, ErrMessageIdNotSupported
	}

	var bodyBytes []byte
	if header.Attr.SegmentationEnabled {
		bodyBytes = unescapedData[MessageHeaderMaxLength : len(unescapedData)-1]
	} else {
		bodyBytes = unescapedData[MessageHeaderNormalLength : len(unescapedData)-1]
	}

	body, err := c.body.Decode(bodyBytes)
	if err != nil {
		return nil, err
	}
	msg.B = body

	return &msg, nil
}
