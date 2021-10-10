/*
A client example
*/
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"time"
)

var host string
var port int

func init() {
	flag.StringVar(&host, "h", "127.0.0.1", "host, default 127.0.0.1")
	flag.IntVar(&port, "p", 9090, "port, default 9090")
}

func main() {
	flag.Parse()

	var str = "7e02000149019161017001000000000000000040000158708c06c94a6e00000000000016101710275654470aec26cad75fdec1dc9c9fcdf89cbb9c216ade77b2b8c83a354e4ab8b5388345ac2af4b31cfa6883aafcb3a42940641e5db1b0411d0abae2aef8dfa8f07d0140adec26ca1986e6adef7be60a0601cc000024900e6100000000ffaa00000000000001cc000024900e6d00000000ffae00000000000001cc00002490128600000000ffa300000000000001cc000024900e6b00000000ffa100000000000001cc000024900ffd00000000ff9b00000000000001cc00002490114500000000ff9a000000000000fe65e602000162f2000c000151800100000000000000f3000102f400010ef5000100f900040000063520000a898602b513165013127007002e563a392e302e3030305432323b4353513a31342c302c312c312c302c322c302c302c313031383130303935312c303a7e"

	data, _ := hex.DecodeString(str)

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(host), Port: port})
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		var allDataBytes []byte
		var buff = make([]byte, 1)
		var readStartTag = false
		for {
			_, err := conn.Read(buff)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("EOF: %s\n", err)
					return
				}
				fmt.Printf("read failed: %s\n", err)
				return
			}
			if !readStartTag {
				// catch start
				if buff[0] == 0x7e {
					readStartTag = true
				} else {
					continue
				}
			} else {
				// catch end
				if buff[0] == 0x7e {
					fmt.Printf("received jtt808 message: 7e%s7e\n", hex.EncodeToString(allDataBytes))
					allDataBytes = allDataBytes[:0]
					readStartTag = false
					continue
				}
				allDataBytes = append(allDataBytes, buff[0])
			}
		}
	}()

	for {
		fmt.Println("sending new data...")
		conn.Write(data)
		time.Sleep(time.Second)
	}
}
