package protocol

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/sceneryback/jtt808/log"
)

const (
	MessageHeaderMaxLength    = 16
	MessageHeaderNormalLength = 12
)

var logger = log.Logger

type BodyAttr struct {
	SegmentationEnabled bool   `json:"是否分包,omitempty"`
	Preserved           uint8  `json:"保留字段,omitempty"`
	EncryptionMethod    string `json:"数据加密方式,omitempty"`
	BodyLength          uint16 `json:"消息体长度,omitempty"`
}

type SegmentInfo struct {
	TotalSegments uint16 `json:"消息包总数,omitempty"`
	SegmentNum    uint16 `json:"包序号,omitempty"`
}

type Header struct {
	MessageId uint16       `json:"消息 ID,omitempty"`
	Attr      *BodyAttr    `json:"消息体属性,omitempty"`
	Phone     uint64       `json:"终端手机号,omitempty"`
	SerialNum uint16       `json:"消息流水号,omitempty"`
	SegInfo   *SegmentInfo `json:"消息包封装项,omitempty"`
}

type Body interface {
	Data() []byte
}

type Message struct {
	Identifier uint8   `json:"标识位,omitempty"`
	H          *Header `json:"消息头,omitempty"`
	B          Body    `json:"消息体,omitempty"`
	Checksum   uint8   `json:"检验码,omitempty"`
}

func (m *Message) Human() (string, error) {
	msgBytes, err := jsoniter.ConfigFastest.MarshalIndent(m, "", "  ")
	if err != nil {
		logger.Errorf("failed to marshal message: %s", err)
		return "", err
	}
	return string(msgBytes), err
}
