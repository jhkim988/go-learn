package main

import "fmt"

func main() {
	var StudentsCount [10]int

	for i := 0; i < 10; i++ {
		StudentsCount[i] = i + 1
		fmt.Println(StudentsCount[i])
	}

	var EvenNum [5]int
	EvenNum[0] = 0
	EvenNum[1] = 2
	EvenNum[2] = 4
	EvenNum[3] = 6
	EvenNum[4] = 8

	fmt.Println(EvenNum[2])

	EvenNum2 := [5]int{0, 2, 4, 6, 8}
	fmt.Println(EvenNum2[2])

	for idx, value := range EvenNum {
		fmt.Println(idx, value)
	}

	// 슬라이스는 기존 배열의 한 영역을 나타낸다. 값을 변경하면 기존 배열의 해당 요소가 수정된다.
	numSlice := []int{5, 4, 3, 2, 1}
	sliced := numSlice[3:5]
	fmt.Println(sliced)
	sliced = numSlice[0:]
	fmt.Println(sliced)

	slice2 := make([]int, 5, 10) // 타입, 길이, 용량. 길이는 용량까지 늘어날 수 있다.
	fmt.Println(slice2)
	copy(slice2, numSlice)
	fmt.Println(slice2)

	slice3 := append(numSlice, 3, 0, -1)
	fmt.Println(slice3)

	// 슬라이스 리터럴은 길이가 없는 배열 리터럴과 같다.
	// 배열이 생성되고, 이를 참조하는 슬라이스가 만들어진다.
	sliceLiteral := []bool{true, true, false}
	fmt.Println(sliceLiteral)
}
