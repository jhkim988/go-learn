package main

import (
	"fmt"
	"sync"
	"time"
)

func say(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func sum(s []int, c chan int) {
	sum := 0
	for _, val := range s {
		sum += val
	}
	c <- sum
}

func fibonacci(n int, c chan int) {
	x, y := 0, 1
	for i := 0; i < n; i++ {
		c <- x
		x, y = y, x+y
	}
	close(c)
}

func fibonacciSelect(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}

type SafeCounter struct {
	mu sync.Mutex
	v  map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	c.v[key]++
	c.mu.Unlock()
}

func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v[key]
}

func main() {
	// goroutine: Go 런타임에 의해 관리되는 경량 스레드
	// go 키워드를 이용하여 고루틴을 실행한다.
	// 함수의 실행은 새로운 고루틴에서 일어난다.
	go say("world")
	say("hello")

	// 채널: 채널연산자(<-)를 통해 값을 주고 받을 수 있는 분리된 통로
	// ch <- v : 채널 ch 에 값 v 를 전송한다.
	// v := <- ch : 채널 ch 로부터 값을 받고 v 에 대입한다.
	// ch := make(chan int) : make 연산자로 채널을 만든다.
	// 전송과 수신은 다른 한쪽이 준비될 때까지 block 된다.
	s := []int{7, 2, 8, -9, 4, 0}
	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c
	fmt.Println(x, y, x+y)

	// Buffered Channels
	// 버퍼채널로 전송: 크기가 꽉차면 block 된다.
	// 버퍼 채널로부터 수신: 크기가 0이면 block 된다.
	chbuff := make(chan int, 100)
	chbuff <- 1
	chbuff <- 2
	fmt.Println(<-chbuff)
	fmt.Println(<-chbuff)

	// close: 전송하는 쪽에서 더 이상 보낼 데이터가 없다는 것을 암시하기 위해 channel 을 닫는다.
	// v, ok := <- c // close 되면 ok 가 false 다.
	// 반드시 전송자만 channel 을 닫아야 하고, 닫힌 channel 에 전송하는 것은 panic 이 발생된다.
	// 보통 channel 을 닫을 필요는 없다. 수신자가 (반복문 등에서) 더 이상 들어오는 값이 없다는 걸 알아야 하는 경우에만 닫는다.
	c = make(chan int, 10)
	go fibonacci(cap(c), c)
	for i := range c {
		fmt.Println(i)
	}

	// select: 다중 커뮤니케이션 연산에서 대기할 수 있게 한다.
	// case 중 하나가 실행될 때까지 block 되고 해당하는 case 를 수행한다.
	// 다수의 case 가 준비되는 경우 무작위 하나가 실행된다.
	// default select: 다른 case 들이 모두 준비되지 않았을 때 실행된다.
	c = make(chan int)
	quit := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<-c)
		}
		quit <- 0
	}()
	fibonacciSelect(c, quit)

	// sync.Mutex
	// Lock, Unlock 메서드를 이용해 mutual exclusion 을 이용할 수 있다.
	counter := SafeCounter{v: make(map[string]int)}
	for i := 0; i < 1000; i++ {
		go counter.Inc("someKey")
	}
	time.Sleep(time.Second)
	fmt.Println(counter.Value("someKey"))
}
