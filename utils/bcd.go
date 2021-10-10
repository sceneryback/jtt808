package utils

import (
	"encoding/hex"
	"strconv"
)

// 12345
func EncodeBCD(s string) []byte {
	bs, _ := hex.DecodeString(s)
	return bs
}

// 0x161017102756  --  时间，6 字节，8421 码，即 2016-10-17 10:27:56
func DecodeBCD(data []byte) string {
	var result string

	var part0, part1 byte
	for i := range data {
		part0 = (data[i] & 0xf0) >> 4
		part1 = data[i] & 0x0f
		result += strconv.Itoa(int(part0))
		result += strconv.Itoa(int(part1))
	}

	return result
}
