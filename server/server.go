/*
A server example
*/
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"

	jtt808 "github.com/sceneryback/jtt808/codec"
)

var port int
var codec jtt808.Codec

func init() {
	flag.IntVar(&port, "p", 9090, "tcp port, default 9090")

	codec, _ = jtt808.NewCodec(nil)
}

func handleSingleMessage(conn net.Conn, data []byte) {
	var resp = jtt808.Message{
		H: &jtt808.Header{
			MessageId: 0x8001,
			Attr: &jtt808.BodyAttr{
				BodyLength: 5,
			},
		},
	}
	var respBody = jtt808.ServerResponse{}

	defer func() {
		resp.B = &respBody
		respBytes, _ := codec.Encode(&resp)
		conn.Write(respBytes)
	}()

	msg, err := codec.Decode(data)

	if msg != nil && msg.H != nil {
		respBody.ID = msg.H.MessageId
		respBody.SerialNum = msg.H.SerialNum
	}

	if err != nil {
		fmt.Printf("failed to decode: %s\n", err)
		respBody.Result = 0x01
		return
	}
	fmt.Println(msg.Human())

	respBody.Result = 0x0
}

func serveConn(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("received conn from %s\n", conn.RemoteAddr().String())

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
				go handleSingleMessage(conn, allDataBytes)
				allDataBytes = allDataBytes[:0]
				readStartTag = false
				continue
			}
			allDataBytes = append(allDataBytes, buff[0])
		}
	}
}

func main() {
	flag.Parse()

	ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	if err != nil {
		fmt.Println("failed to listen tcp:", err)
		return
	}
	fmt.Println("listen on :",port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go serveConn(conn)
	}
}
