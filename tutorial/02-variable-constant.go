package main

import "fmt"

func main() {
	var a int = 5
	var b float32 = 4.32
	const pi float64 = 3.141592
	var (
		c = 8
		d = 7
	)
	x, y := 14, 15

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(pi)
	fmt.Println(c)
	fmt.Println(d)
	fmt.Println(x, ",", y)
}
