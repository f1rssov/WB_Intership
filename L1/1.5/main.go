package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	n := 5 * time.Second

	go func() {
		i := 1
		for {
			ch <- i
			i++
			time.Sleep(500 * time.Millisecond)
		}
	}()

	stop := time.After(n)

	for {
		select {
		case v := <-ch:
			fmt.Println(v)
		case <-stop:
			fmt.Println("stop")
			return
		}
	}
}
