package codec

import (
	"bytes"
	"fmt"
)

const (
	MessageHeaderMaxLength    = 16
	MessageHeaderNormalLength = 12
)

type BodyAttr struct {
	SegmentationEnabled bool
	Preserved           uint8
	EncryptionMethod    string
	BodyLength          uint16
}

func (b *BodyAttr) Human() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("segmentation enabled: %v\n", b.SegmentationEnabled))
	buf.WriteString(fmt.Sprintf("preserved: %d\n", b.Preserved))
	buf.WriteString(fmt.Sprintf("encryption method: %s\n", b.EncryptionMethod))
	buf.WriteString(fmt.Sprintf("body length: %d\n", b.BodyLength))

	return buf.String()
}

type SegmentInfo struct {
	TotalSegments uint16
	SegmentNum    uint16
}

type Header struct {
	MessageId uint16
	Attr      *BodyAttr
	Phone     uint64
	SerialNum uint16
	SegInfo   *SegmentInfo
}

func (h *Header) Human() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("message id: %d\n", h.MessageId))
	buf.WriteString(fmt.Sprintf("attribute: %s\n", h.Attr.Human()))
	buf.WriteString(fmt.Sprintf("phone: %d\n", h.Phone))
	buf.WriteString(fmt.Sprintf("serial num: %d\n", h.SerialNum))
	if h.Attr.SegmentationEnabled {
		buf.WriteString(fmt.Sprintf("total segments: %d\n", h.SegInfo.TotalSegments))
		buf.WriteString(fmt.Sprintf("segment num: %d\n", h.SegInfo.SegmentNum))
	}

	return buf.String()
}

type Body interface {
	Human() string
}

type Message struct {
	Identifier uint8
	H          *Header
	B          Body
	Checksum   uint8
}

func (m *Message) Human() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("identifier: %d\n", m.Identifier))
	buf.WriteString(fmt.Sprintf("header:\n%s\n", m.H.Human()))
	buf.WriteString(fmt.Sprintf("body:\n%s\n", m.B.Human()))

	return buf.String()
}

type ServerResponse struct {
	SerialNum uint16
	ID        uint16
	Result    uint8
}

func (r *ServerResponse) Human() string {
	var buf bytes.Buffer

	var res string
	if r.Result == 1 {
		res = "failure"
	} else if r.Result == 0 {
		res = "success"
	}

	buf.WriteString(fmt.Sprintf("serial num: %d\n", r.SerialNum))
	buf.WriteString(fmt.Sprintf("message id: %d\n", r.ID))
	buf.WriteString(fmt.Sprintf("result: %s\n", res))

	return buf.String()
}
