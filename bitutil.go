package main

import (
	"fmt"
)

func bitsInByte(b byte) []byte {
	fmt.Printf("convert %d\n", b)
	res := make([]byte, 8, 8)
	for i := uint8(0); i < 8; i++ {
		res[i] = b >> (7 - i) & 0x1
	}
	return res
}
