package codec

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/sceneryback/jtt808/codec/location"
	"github.com/sceneryback/jtt808/codec/protocol"
	"github.com/sceneryback/jtt808/codec/response"
	"github.com/sceneryback/jtt808/log"
)

var (
	ErrChecksumFailed        = errors.New("failed to verify checksum")
	ErrDecodeHeaderFailed    = errors.New("failed to decode header")
	ErrMessageIdNotSupported = errors.New("message id not supported yet")
)

var logger = log.Logger

type Codec interface {
	Encode(*protocol.Message) ([]byte, error)
	Decode([]byte) (*protocol.Message, error)
}

type HeaderCodec interface {
	Encode(header *protocol.Header) ([]byte, error)
	Decode([]byte) (*protocol.Header, error)
}

type BodyCodec interface {
	Encode(protocol.Body) ([]byte, error)
	Decode([]byte) (protocol.Body, error)
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

func (c *codec) Encode(msg *protocol.Message) ([]byte, error) {
	var buf bytes.Buffer

	headerBytes, err := c.header.Encode(msg.H)
	if err != nil {
		logger.Errorf("failed to encode header: %s", err)
		return nil, err
	}

	buf.Write(headerBytes)

	switch msg.H.MessageId {
	case 0x8001:
		c.body = response.NewServerResponseCodec()
	default:
		logger.Errorf("message id not supported: %s", msg.H.MessageId)
		return nil, ErrMessageIdNotSupported
	}

	bodyBytes, err := c.body.Encode(msg.B)
	if err != nil {
		logger.Errorf("failed to encode body: %s", err)
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
func (c *codec) checksumVerified(unescapedMsg []byte) (uint8, bool) {
	unescapedLength := len(unescapedMsg)

	var checksum = unescapedMsg[unescapedLength-1:][0]
	realsum := c.checksum(unescapedMsg[0 : unescapedLength-1])

	return checksum, realsum == checksum
}

func (c *codec) decodeHeader(h []byte) (*protocol.Header, error) {
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

	err = binary.Read(bytes.NewReader(phoneBytes), binary.BigEndian, &header.Phone)
	if err != nil {
		logger.Errorf("failed to read phone: %s", err)
		return nil, err
	}

	err = binary.Read(bytes.NewReader(serialNumBytes), binary.BigEndian, &header.SerialNum)
	if err != nil {
		logger.Errorf("failed to read serial number: %s", err)
		return nil, err
	}

	header.Attr = &protocol.BodyAttr{}
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

func (c *codec) trimIdentifiers(data []byte) []byte {
	if data[0] == 0x7e {
		data = data[1:]
	}
	if data[len(data)-1] == 0x7e {
		data = data[:len(data)-1]
	}
	return data
}

func (c *codec) Decode(data []byte) (*protocol.Message, error) {
	data = c.trimIdentifiers(data)

	unescapedData := c.unescape(data)

	checksum, valid := c.checksumVerified(unescapedData)
	if !valid {
		logger.Error("checksum failed")
		return nil, ErrChecksumFailed
	}

	var msg protocol.Message

	header, err := c.header.Decode(unescapedData[:protocol.MessageHeaderMaxLength])
	if err != nil {
		logger.Errorf("failed to decode header: %s", err)
		return nil, ErrDecodeHeaderFailed
	}
	msg.H = header

	switch msg.H.MessageId {
	case 0x0001:
		c.body = response.NewClientResponseCodec()
	case 0x8001:
		c.body = response.NewServerResponseCodec()
	// case 0x0002:
	// 	c.body = &clientHeartbeatCodec{}
	// case 0x0100:
	// 	c.body = &clientRegisterCodec{}
	// case 0x8100:
	// 	c.body = &clientRegisterResponseCodec{}
	// case 0x0003:
	// 	c.body = &clientLogoutCodec{}
	// case 0x0102:
	// 	c.body = &clientAuthorizationCodec{}
	// case 0x8103:
	// 	c.body = &setClientParameterCodec{}
	// case 0x8104:
	// 	c.body = &lookupClientParameterCodec{}
	// case 0x0104:
	// 	c.body = &lookupClientParameterResponseCodec{}
	// case 0x8105:
	// 	c.body = &clientControlCodec{}
	case 0x0200:
		c.body = location.NewLocationCodec()
	// case 0x8201:
	// 	c.body = &locationLookupCodec{}
	// case 0x0201:
	// 	c.body = &locationLookupResponseCodec{}
	// case 0x8202:
	// 	c.body = &tempLocationTrackingCodec{}
	// case 0x8300:
	// 	c.body = &textDeliverCodec{}
	// case 0x8301:
	// 	c.body = &eventSettingCodec{}
	// case 0x0301:
	// 	c.body = &eventReportCodec{}
	// case 0x8302:
	// 	c.body = &queryDeliverCodec{}
	// case 0x0302:
	// 	c.body = &queryAnswerCodec{}
	// case 0x8303:
	// 	c.body = &infoOnDemandMenuSettingCodec{}
	// case 0x0303:
	// 	c.body = &infoOnDemandPlayCancelCodec{}
	// case 0x8304:
	// 	c.body = &infoServiceCodec{}
	// case 0x8400:
	// 	c.body = &callBackCodec{}
	// case 0x8401:
	// 	c.body = &phoneBookSettingCodec{}
	// case 0x8500:
	// 	c.body = &vehicleControlCodec{}
	// case 0x0500:
	// 	c.body = &vehicleControlResponseCodec{}
	// case 0x8600:
	// 	c.body = &circleSettingCodec{}
	// case 0x8601:
	// 	c.body = &circleDeleteCodec{}
	// case 0x8602:
	// 	c.body = &rectangleSettingCodec{}
	// case 0x8603:
	// 	c.body = &rectangleDeleteCodec{}
	// case 0x8604:
	// 	c.body = &polygonSettingCodec{}
	// case 0x8605:
	// 	c.body = &polygonDeleteCodec{}
	// case 0x8606:
	// 	c.body = &routeSettingCodec{}
	// case 0x8607:
	// 	c.body = &routeDeleteCodec{}
	// case 0x8700:
	// 	c.body = &drivingRecordsCollectCodec{}
	// case 0x0700:
	// 	c.body = &drivingRecordsUploadCodec{}
	// case 0x8701:
	// 	c.body = &drivingRecordsParameterDownloadCodec{}
	// case 0x0701:
	// 	c.body = &digitalWayBillCodec{}
	// case 0x0702:
	// 	c.body = &driverIdentityReportCodec{}
	// case 0x0800:
	// 	c.body = &multiMediaEventReportCodec{}
	// case 0x0801:
	// 	c.body = &multiMediaDataReportCodec{}
	// case 0x8800:
	// 	c.body = &multiMediaDataReportResponseCodec{}
	// case 0x8801:
	// 	c.body = &cameraShotCommandCodec{}
	// case 0x8802:
	// 	c.body = &savedMultiMediaDataRetrivalCodec{}
	// case 0x0802:
	// 	c.body = &savedMultiMediaDataRetrivalResponseCodec{}
	// case 0x8803:
	// 	c.body = &savedMultiMediaDataReportCodec{}
	// case 0x8804:
	// 	c.body = &startRecordingCodec{}
	// case 0x8900:
	// 	c.body = &dataDownPassThroughCodec{}
	// case 0x0900:
	// 	c.body = &dataUpPassThroughCodec{}
	// case 0x0901:
	// 	c.body = &dataCompressUploadCodec{}
	// case 0x8A00:
	// 	c.body = &rsaKeyCodec{}
	// case 0x0A00:
	// 	c.body = &deviceRsaKeyCodec{}
	default:
		logger.Error("message id not supported: %s", msg.H.MessageId)
		return nil, ErrMessageIdNotSupported
	}

	var bodyBytes []byte
	if header.Attr.SegmentationEnabled {
		bodyBytes = unescapedData[protocol.MessageHeaderMaxLength : len(unescapedData)-1]
	} else {
		bodyBytes = unescapedData[protocol.MessageHeaderNormalLength : len(unescapedData)-1]
	}

	body, err := c.body.Decode(bodyBytes)
	if err != nil {
		logger.Errorf("failed to decode body: %s", err)
		return nil, err
	}
	msg.B = body

	msg.Identifier = 0x7e
	msg.Checksum = checksum

	return &msg, nil
}
