package main

import (
	"fmt"
	"strings"
)

// WellAlignedStruct demonstrates proper struct field alignment.
type WellAlignedStruct struct {
	A int64 // 8 bytes
	B int64 // 8 bytes
	C int32 // 4 bytes
	D int16 // 2 bytes
	E int8  // 1 byte
	F bool  // 1 byte
}

func main() {
	// Pre-allocated slice — best practice
	n := 100
	items := make([]int, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, i)
	}

	// strings.Builder — best practice for string concatenation
	var sb strings.Builder
	for i := 0; i < 10; i++ {
		sb.WriteString("hello ")
	}

	s := WellAlignedStruct{A: 1, B: 2, C: 3, D: 4, E: 5, F: true}
	fmt.Println(len(items), sb.String(), s)
}
