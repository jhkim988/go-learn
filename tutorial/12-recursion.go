package main

import "fmt"

func main() {
	num := 5
	fmt.Println(factorial(num))
}

func factorial(x int) int {
	if x == 0 {
		return 1
	}
	return x * factorial(x-1)
}
