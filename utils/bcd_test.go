package utils

import (
	"fmt"
	"testing"
)

func TestEncodeBCD(t *testing.T) {
	s := "012345678901"
	bs := EncodeBCD(s)
	fmt.Printf("%x\n", bs)
}
