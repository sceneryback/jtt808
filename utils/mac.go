package utils

import "net"

func MacIntegerToStr(n int64) string {
	hwaddr := make(net.HardwareAddr, 6)
	hwaddr[0] = byte(n >> 40)
	hwaddr[1] = byte(n >> 32)
	hwaddr[2] = byte(n >> 24)
	hwaddr[3] = byte(n >> 16)
	hwaddr[4] = byte(n >> 8)
	hwaddr[5] = byte(n)
	return hwaddr.String()
}

func HexBytesToMacAddr(src []byte) string {
	return net.HardwareAddr(src).String()
}
