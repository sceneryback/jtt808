package response

import (
	"bytes"
	"fmt"
)

type ServerResponse struct {
	SerialNum uint16
	ID        uint16
	Result    uint8

	data []byte
}

func (r *ServerResponse) Data() []byte {
	return r.data
}

// TODO: consider use jsoniter
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

type ClientResponse struct {
	SerialNum uint16
	ID        uint16
	Result    uint8

	data []byte
}

func (r *ClientResponse) Data() []byte {
	return r.data
}

func (r *ClientResponse) Human() string {
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
