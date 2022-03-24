package location

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"

	"github.com/sceneryback/jtt808/codec/protocol"
	"github.com/sceneryback/jtt808/log"
	"github.com/sceneryback/jtt808/utils"
)

const (
	LocationBasicInfoLength = 28

	TimeFormat = "20060102150405-0700"

	TimeFormatHuman = "2006-01-02 15:04:05"
)

var logger = log.Logger

type locationBasicInfoCodec struct {
}

type locationAdditionalInfoCodec struct {
}

type locationCodec struct {
	basic locationBasicInfoCodec
	ai    locationAdditionalInfoCodec
}

func NewLocationCodec() *locationCodec {
	return &locationCodec{}
}

func (l *locationBasicInfoCodec) Decode(data []byte) (*BasicInfo, error) {
	var basic BasicInfo

	err := binary.Read(bytes.NewReader(data[:4]), binary.BigEndian, &basic.Alert)
	if err != nil {
		logger.Errorf("failed to read alert: %s", err)
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[4:8]), binary.BigEndian, &basic.State)
	if err != nil {
		logger.Errorf("failed to read state: %s", err)
		return nil, err
	}

	var lat uint32
	err = binary.Read(bytes.NewReader(data[8:12]), binary.BigEndian, &lat)
	if err != nil {
		logger.Errorf("failed to read latitude: %s", err)
		return nil, err
	}
	basic.Latitude = float32(lat) / 1e6

	var lon uint32
	err = binary.Read(bytes.NewReader(data[12:16]), binary.BigEndian, &lon)
	if err != nil {
		logger.Errorf("failed to read longitude: %s", err)
		return nil, err
	}
	basic.Longitude = float32(lon) / 1e6

	err = binary.Read(bytes.NewReader(data[16:18]), binary.BigEndian, &basic.Altitude)
	if err != nil {
		logger.Errorf("failed to read altitude: %s", err)
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[18:20]), binary.BigEndian, &basic.Speed)
	if err != nil {
		logger.Errorf("failed to read speed: %s", err)
		return nil, err
	}

	err = binary.Read(bytes.NewReader(data[20:22]), binary.BigEndian, &basic.Direction)
	if err != nil {
		logger.Errorf("failed to read direction: %s", err)
		return nil, err
	}

	tstr := strings.TrimPrefix(utils.DecodeBCD(data[22:]), "0")
	ts, err := time.Parse(TimeFormat, "20"+tstr+"+0800")
	if err != nil {
		logger.Errorf("failed to read time: %s", err)
		return nil, err
	}
	basic.Timestamp = ts.Format(TimeFormatHuman)

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
				logger.Errorf("failed to decode wifi: %s", err)
				return nil, err
			}
			infos = append(infos, wifis)
			i += (2 + singleInfoLength)
		case 0x56:
			singleInfoLength = int(data[i+1])
			batInfo, err := (&batteryCodec{}).Decode(data[i+2 : i+2+singleInfoLength])
			if err != nil {
				logger.Errorf("failed to decode battery info: %s", err)
				return nil, err
			}
			infos = append(infos, batInfo)
			i += (2 + singleInfoLength)
		// TODO: other location infos
		default:
			// ignored other infos
			singleInfoLength = int(data[i+1])
			info, err := (&unknownInfoCodec{}).Decode(data[i : i+1+singleInfoLength+1])
			if err != nil {
				logger.Errorf("failed to decode unknown info: %s", err)
				return nil, err
			}
			infos = append(infos, info)
			i += (2 + singleInfoLength)
		}
	}

	return infos, nil
}

func (l *locationCodec) Encode(b protocol.Body) ([]byte, error) {
	return nil, nil
}

func (l *locationCodec) Decode(data []byte) (protocol.Body, error) {
	var body LocationMsgBody

	basicBytes := data[:LocationBasicInfoLength]
	basic, err := l.basic.Decode(basicBytes)
	if err != nil {
		logger.Errorf("failed to decode basic: %s", err)
		return nil, err
	}
	body.Basic = basic
	body.BasicHex = hex.EncodeToString(basicBytes)

	additionalBytes := data[LocationBasicInfoLength:]
	additionals, err := l.ai.Decode(additionalBytes)
	if err != nil {
		logger.Errorf("failed to decode additional info: %s", err)
		return nil, err
	}
	body.AdditionalInfos = additionals
	body.AdditionalInfosHex = hex.EncodeToString(additionalBytes)

	body.data = data

	return &body, nil
}
