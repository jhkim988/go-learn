package main

import (
	"fmt"
	"math"
)

func main() {
	rect1 := Rectangle{height: 10, width: 5}
	rect2 := Rectangle{10, 5}
	fmt.Println(rect1.height)
	fmt.Println(rect2)

	fmt.Println("Area of rectangle is ", rect1.area())

	circle := Circle{7}
	fmt.Println("Area of circle is ", getArea(circle))
}

type Rectangle struct {
	height float64
	width  float64
}

// Rectangle 구조체에 대한 area 함수 선언
// Receiver - 구조체 참조에 대한 인자
func (rect Rectangle) area() float64 {
	return rect.height * rect.width
}

type Circle struct {
	radius float64
}

func (c Circle) area() float64 {
	return math.Pi * c.radius * c.radius
}

type Shape interface {
	area() float64
}

func getArea(shape Shape) float64 {
	return shape.area()
}
