package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	file, err := os.Create("sample.txt")

	if err != nil {
		log.Fatal(err)
	}

	file.WriteString("Hello File IO")
	file.Close()

	stream, err := ioutil.ReadFile("sample.txt")

	if err != nil {
		log.Fatal(err)
	}

	s1 := string(stream)
	fmt.Println(s1)
}
