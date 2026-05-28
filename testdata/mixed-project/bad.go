package main

import "fmt"

func badFunction() {
	// Append in loop without pre-allocation (triggers CEG-016)
	var data []string
	for i := 0; i < 200; i++ {
		data = append(data, "item")
	}
	fmt.Println(len(data))
}
