package main

import "fmt"

func main() {
	z := 10
	fmt.Println(z)
	fmt.Println(&z)

	changeValue(&z)
	fmt.Println(z)
}

func changeValue(z *int) {
	*z = 7
}
