package main

import "fmt"

func main() {
	const pi float64 = 3.141592
	x := 5
	isbool := true

	fmt.Printf("%f \n", pi)
	fmt.Printf("%.3f \n", pi)   // 반올림
	fmt.Printf("%T \n", isbool) // T: type
	fmt.Printf("%t \n", pi)     // t: bool, wrong
	fmt.Printf("%t \n", isbool)
	fmt.Printf("%d \n", x)
	fmt.Printf("%b \n", 5)  // binary
	fmt.Printf("%c \n", 33) // char, ascii
	fmt.Printf("%x \n", 15) // hex
	fmt.Printf("%e \n", pi)

	var name string = "jhkim"
	fmt.Println(len(name))
	fmt.Println(name + "-stjohn")

}
