package main

import (
	// "flag"
	"fmt"
)

// func testmain() {
// 	// byteToWrite := flag.Int("b", 200, "byte to send over spi")
// 	virusZip := flag.String("z", "", "virus to kill AI")
// 	flag.Parse()
// 	if nil != virusZip {
// 		fmt.Printf("uploading virusfile %s\n", *virusZip)
// 	}
// 	go func() {
// 		for v := range channel {
// 			fmt.Printf("content (length %d) %v\n", len(v), v)
// 		}
// 	}()
// 	stream(*virusZip)
// }

func bitsInByte(b byte) []byte {
	res := make([]byte, 8, 8)
	for i := uint8(0); i < 8; i++ {
		res[i] = b >> (7 - i) & 0x1
	}
	fmt.Printf("converted %d->%v\n", b, res)
	return res
}
