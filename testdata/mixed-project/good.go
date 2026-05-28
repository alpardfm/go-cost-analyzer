package main

import "fmt"

func goodFunction() {
	// Pre-allocated slice — best practice
	n := 50
	items := make([]int, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, i)
	}
	fmt.Println(len(items))
}
