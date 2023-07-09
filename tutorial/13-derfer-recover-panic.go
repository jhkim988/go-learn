package main

import "fmt"

func main() {
	// 1. defer 키워드 - 자신을 둘러싼 함수가 종료할 때까지 실행을 연기한다.
	// 연기된 함수들은 스택에 쌓이고, 후입선출로 수행된다.
	// defer FirstRun()
	// SecondRun()

	// 2. 0으로 나누어 에러를 발생 시킨다. 에러가 발생해도 defer 는 실행이 보장된다.
	// recover 를 통해 에러/panic 이 발생해도 회복할 수 있다. recover 는 panic 을 호출할 때 인자를 반환한다.
	fmt.Println("div(3, 0) = ", div(3, 0))
	fmt.Println("div(3, 5) = ", div(3, 5))
	demPanic()

	// 3. 무제한의 인자를 전달하는 방법
	fmt.Println(addemup(10, 20, 30, 40, 50))
}

func FirstRun() {
	fmt.Println("I executed First")
}

func SecondRun() {
	fmt.Println("I executed Second")
}

func div(num1, num2 int) int {
	defer func() {
		fmt.Println("div recover", recover())
	}()

	solution := num1 / num2
	return solution
}

func demPanic() {
	defer func() {
		fmt.Println("demPanic recover", recover())
	}()

	panic("Panic")
}

func addemup(args ...int) int {
	sum := 0
	for _, value := range args {
		sum += value
	}
	return sum
}
