package main

import "fmt"

func main() {
	for i := 0; i < 5; i++ {
		fmt.Println(i)
	}

	i := 1
	for i <= 10 {
		fmt.Println(i)
		i++
	}

	for i := 1; i < 5; i++ {
		for j := 1; j < 5; j++ {
			fmt.Println(i, ", ", j)
		}
	}
}
