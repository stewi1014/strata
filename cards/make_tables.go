//go:build ignore

package main

import (
	"fmt"
	"math/bits"
	"os"
	"strings"
)

func main() {
	buff := new(strings.Builder)
	buff.WriteString("package cards\n")

	pop4table(buff)
	rankers8table(buff)

	err := os.WriteFile("cards_tables.go", []byte(buff.String()), 0666)
	if err != nil {
		panic(err)
	}
}

func pop4(n int) byte {
	var b byte

	for n != 0 {
		n &= n - 1
		b++
	}

	return b
}

func pop4table(buff *strings.Builder) {
	buff.WriteString("\nconst pop4table = \"")
	for i := range 0x10 {
		fmt.Fprintf(buff, "\\x%02x", pop4(i))
	}
	buff.WriteString("\"\n")
}

func rankers8(n int) byte {
	var ret byte
	if n&0xF0 > 0 {
		ret |= 1 << bits.TrailingZeros(uint(n&0xF0))
	}

	if n&0x0F > 0 {
		ret |= 1 << bits.TrailingZeros(uint(n&0x0F))
	}

	return ret
}

func rankers8table(buff *strings.Builder) {
	buff.WriteString("\nconst rankers8table = \"")
	for i := range 0x100 {
		if i%16 == 0 {
			buff.WriteString("\" +\n\t\"")
		}
		fmt.Fprintf(buff, "\\x%02x", rankers8(i))
	}
	buff.WriteString("\"\n")
}
