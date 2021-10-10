package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sceneryback/jtt808/utils"
	"strings"
	"time"
)

const (
	LocationBasicInfoLength = 28

	TimeFormat = "20060102150405-0700"

	TimeFormatHuman = "2006-01-02 15:04:05"
)

type locationBasicInfoCodec struct {
}

type locationAdditionalInfoCodec struct {
}

type locationCodec struct {
	basic locationBasicInfoCodec
	ai    locationAdditionalInfoCodec
}

type LocationAdditionalInfo interface {
	Id() uint8
	Length() uint8
	Info() []byte
	Human() string
}

type BasicInfo struct {
	Alert     uint32
	State     uint32
	Latitude  uint32
	Longitude uint32
	Altitude  uint16
	Speed     uint16
	Direction uint16
	Timestamp int64
}

func (l *locationBasicInfoCodec) Decode(data []byte) (*BasicInfo, error) {
	var basic BasicInfo

	err := binary.Read(bytes.NewReader(data[:4]), binary.BigEndian, &basic.Alert)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[4:8]), binary.BigEndian, &basic.State)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[8:12]), binary.BigEndian, &basic.Latitude)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[12:16]), binary.BigEndian, &basic.Longitude)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[16:18]), binary.BigEndian, &basic.Altitude)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[18:20]), binary.BigEndian, &basic.Speed)
	if err != nil {
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[20:22]), binary.BigEndian, &basic.Direction)
	if err != nil {
		return nil, err
	}

	tstr := strings.TrimPrefix(utils.DecodeBCD(data[22:]), "0")
	ts, err := time.Parse(TimeFormat, "20"+tstr+"+0800")
	if err != nil {
		return nil, err
	}
	basic.Timestamp = ts.Unix()

	return &basic, nil
}

func (l *locationAdditionalInfoCodec) Decode(data []byte) ([]LocationAdditionalInfo, error) {
	var infos []LocationAdditionalInfo

	var singleInfoLength int
	for i := 0; i < len(data); {
		switch data[i] {
		case 0x54:
			singleInfoLength = int(data[i+1])
			wifis, err := (&wifiCodec{}).Decode(data[i+2 : i+2+singleInfoLength])
			if err != nil {
				return nil, err
			}
			infos = append(infos, wifis)
			i += (2 + singleInfoLength)
		case 0x56:
			singleInfoLength = int(data[i+1])
			batInfo, err := (&batteryCodec{}).Decode(data[i+2 : i+2+singleInfoLength])
			if err != nil {
				return nil, err
			}
			infos = append(infos, batInfo)
			i += (2 + singleInfoLength)
		default:
			// ignored other infos
			singleInfoLength = int(data[i+1])
			info, err := (&unknownInfoCodec{}).Decode(data[i : i+1+singleInfoLength+1])
			if err != nil {
				return nil, err
			}
			infos = append(infos, info)
			i += (2 + singleInfoLength)
		}
	}

	return infos, nil
}

type LocationMsgBody struct {
	Basic           *BasicInfo
	AdditionalInfos []LocationAdditionalInfo
}

func (l *LocationMsgBody) Human() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("alert: %d\n", l.Basic.Alert))
	buf.WriteString(fmt.Sprintf("state: %d\n", l.Basic.State))
	buf.WriteString(fmt.Sprintf("latitude: %f\n", float64(l.Basic.Latitude)/1e6))
	buf.WriteString(fmt.Sprintf("longitude: %f\n", float64(l.Basic.Longitude)/1e6))
	buf.WriteString(fmt.Sprintf("altitude: %d\n", l.Basic.Altitude))
	buf.WriteString(fmt.Sprintf("speed: %d\n", l.Basic.Speed))
	buf.WriteString(fmt.Sprintf("direction: %d\n", l.Basic.Direction))
	buf.WriteString(fmt.Sprintf("timestamp: %s\n", time.Unix(l.Basic.Timestamp, 0).Format(TimeFormatHuman)))

	buf.WriteString("additional infos ===== \n")
	for i := range l.AdditionalInfos {
		buf.WriteString(fmt.Sprintf("%s\n", l.AdditionalInfos[i].Human()))
	}

	return buf.String()
}

func (l *locationCodec) Encode(b Body) ([]byte, error) {
	return nil, nil
}

func (l *locationCodec) Decode(data []byte) (Body, error) {
	var body LocationMsgBody

	basicBytes := data[:LocationBasicInfoLength]
	basic, err := l.basic.Decode(basicBytes)
	if err != nil {
		return nil, err
	}
	body.Basic = basic

	additionalBytes := data[LocationBasicInfoLength:]
	additionals, err := l.ai.Decode(additionalBytes)
	if err != nil {
		return nil, err
	}
	body.AdditionalInfos = additionals

	return &body, nil
}
