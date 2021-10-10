package codec

import (
	"encoding/hex"
	"fmt"
	"github.com/bmizerany/assert"
	"testing"
)

func TestCodec_Decode(t *testing.T) {
	var c, _ = NewCodec(nil)

	var str = "7e02000149019161017001000000000000000040000158708c06c94a6e00000000000016101710275654470aec26cad75fdec1dc9c9fcdf89cbb9c216ade77b2b8c83a354e4ab8b5388345ac2af4b31cfa6883aafcb3a42940641e5db1b0411d0abae2aef8dfa8f07d0140adec26ca1986e6adef7be60a0601cc000024900e6100000000ffaa00000000000001cc000024900e6d00000000ffae00000000000001cc00002490128600000000ffa300000000000001cc000024900e6b00000000ffa100000000000001cc000024900ffd00000000ff9b00000000000001cc00002490114500000000ff9a000000000000fe65e602000162f2000c000151800100000000000000f3000102f400010ef5000100f900040000063520000a898602b513165013127007002e563a392e302e3030305432323b4353513a31342c302c312c312c302c322c302c302c313031383130303935312c303a7e"

	data, _ := hex.DecodeString(str)
	msg, err := c.Decode(data)
	assert.Equal(t, nil, err)
	fmt.Println(msg.Human())
}

func TestDecodeHeaderMsgAttr(t *testing.T) {
	var str = "0200e408019161017001000000300010"

	data, _ := hex.DecodeString(str)

	header, err := (&headerCodec{}).Decode(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, 512, int(header.MessageId))
	assert.Equal(t, 48, int(header.SegInfo.TotalSegments))
	assert.Equal(t, 16, int(header.SegInfo.SegmentNum))
	assert.Equal(t, 19161017001, int(header.Phone))
	assert.Equal(t, 0, int(header.SerialNum))
	assert.Equal(t, 3, int(header.Attr.Preserved))
	assert.Equal(t, "RSA", header.Attr.EncryptionMethod)
	assert.Equal(t, 8, int(header.Attr.BodyLength))
}
