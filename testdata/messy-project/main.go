package main

import "fmt"

// PoorlyAlignedStruct has suboptimal field ordering (triggers CEG-018).
type PoorlyAlignedStruct struct {
	A bool  // 1 byte
	B int64 // 8 bytes — padding before this field
	C bool  // 1 byte
	D int64 // 8 bytes — padding before this field
	E bool  // 1 byte
	F int32 // 4 bytes — padding before this field
}

func main() {
	// Slice append in loop without pre-allocation (triggers CEG-016)
	var items []int
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}

	// Range-based slice append without pre-allocation (triggers CEG-016)
	source := []int{1, 2, 3, 4, 5}
	var copied []int
	for _, v := range source {
		copied = append(copied, v)
	}

	// String concatenation with + in a loop (triggers CEG-017)
	s := ""
	for i := 0; i < 100; i++ {
		s += "hello "
	}

	p := PoorlyAlignedStruct{A: true, B: 1, C: true, D: 2, E: true, F: 3}
	fmt.Println(len(items), len(copied), s, p)
}
