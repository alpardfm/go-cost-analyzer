package main

import "fmt"

func main() {
	//noinspect:CEG-016
	var items []int
	for i := 0; i < 100; i++ {
		items = append(items, i)
	}
	fmt.Println(len(items))
}
