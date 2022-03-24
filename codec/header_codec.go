package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/sceneryback/jtt808/codec/protocol"
	"github.com/sceneryback/jtt808/utils"
)

type headerCodec struct {
}

func (c *headerCodec) Encode(h *protocol.Header) ([]byte, error) {
	var res []byte

	var msgIdBuf bytes.Buffer
	err := binary.Write(&msgIdBuf, binary.BigEndian, h.MessageId)
	if err != nil {
		logger.Errorf("failed to write message id: %s", err)
		return nil, err
	}
	res = append(res, msgIdBuf.Bytes()...)

	var attr uint16
	attr |= uint16(h.Attr.Preserved) << 14
	if h.Attr.SegmentationEnabled {
		attr |= 0x2000
	}
	if h.Attr.EncryptionMethod == "RSA" {
		attr |= 0x0400
	}
	attr |= h.Attr.BodyLength

	var attrBuf bytes.Buffer
	err = binary.Write(&attrBuf, binary.BigEndian, attr)
	if err != nil {
		logger.Errorf("failed to write attr: %s", err)
		return nil, err
	}
	res = append(res, attrBuf.Bytes()...)

	phoneStr := fmt.Sprintf("%d", h.Phone)
	var fillPhoneStr = phoneStr
	if len(phoneStr) < 12 {
		for i := 0; i < 12-len(phoneStr); i++ {
			fillPhoneStr = "0" + fillPhoneStr
		}
	}
	res = append(res, utils.EncodeBCD(fillPhoneStr)...)

	var serialNumBuf bytes.Buffer
	err = binary.Write(&serialNumBuf, binary.BigEndian, h.SerialNum)
	if err != nil {
		logger.Errorf("failed to write serial number: %s", err)
		return nil, err
	}
	res = append(res, serialNumBuf.Bytes()...)

	if h.Attr.SegmentationEnabled {
		var totalSegsBuf, segSeqBuf bytes.Buffer

		err = binary.Write(&totalSegsBuf, binary.BigEndian, &h.SegInfo.TotalSegments)
		if err != nil {
			logger.Errorf("failed to write total segments: %s", err)
			return nil, err
		}

		err = binary.Write(&segSeqBuf, binary.BigEndian, &h.SegInfo.SegmentNum)
		if err != nil {
			logger.Errorf("failed to write segment number: %s", err)
			return nil, err
		}

		res = append(res, totalSegsBuf.Bytes()...)
		res = append(res, segSeqBuf.Bytes()...)
	}

	return res, nil
}

func (c *headerCodec) Decode(h []byte) (*protocol.Header, error) {
	var header protocol.Header

	var msgIdBytes = h[:2]
	var msgAttrBytes = h[2:4]
	var phoneBytes = h[4:10]
	var serialNumBytes = h[10:12]

	err := binary.Read(bytes.NewReader(msgIdBytes), binary.BigEndian, &header.MessageId)
	if err != nil {
		logger.Errorf("failed to read message id: %s", err)
		return nil, err
	}

	phoneStr := strings.TrimPrefix(utils.DecodeBCD(phoneBytes), "0")
	phone, err := strconv.ParseUint(phoneStr, 10, 64)
	if err != nil {
		logger.Errorf("failed to parse phone %s: %s", phoneStr, err)
		return nil, err
	}
	header.Phone = phone

	err = binary.Read(bytes.NewReader(serialNumBytes), binary.BigEndian, &header.SerialNum)
	if err != nil {
		logger.Errorf("failed to read serial number: %s", err)
		return nil, err
	}

	header.Attr = &protocol.BodyAttr{}
	header.Attr.Preserved = uint8(msgAttrBytes[0] >> 6)
	if (msgAttrBytes[0]&0x20)>>5 == 1 {
		header.Attr.SegmentationEnabled = true
	}
	var encrypt = (msgAttrBytes[0] & 0x1C) >> 2
	if encrypt == 1 {
		header.Attr.EncryptionMethod = "RSA"
	}

	var attr uint16
	err = binary.Read(bytes.NewBuffer(msgAttrBytes), binary.BigEndian, &attr)
	if err != nil {
		logger.Errorf("failed to read message attr: %s", err)
		return nil, err
	}
	header.Attr.BodyLength = attr & 0x03ff

	if header.Attr.SegmentationEnabled {
		var segmentBytes = h[12:]

		header.SegInfo = &protocol.SegmentInfo{}

		err = binary.Read(bytes.NewReader(segmentBytes[:2]), binary.BigEndian, &header.SegInfo.TotalSegments)
		if err != nil {
			logger.Errorf("failed to read total segments: %s", err)
			return nil, err
		}
		err = binary.Read(bytes.NewReader(segmentBytes[2:]), binary.BigEndian, &header.SegInfo.SegmentNum)
		if err != nil {
			logger.Errorf("failed to read segment number: %s", err)
			return nil, err
		}
	}

	return &header, nil
}
