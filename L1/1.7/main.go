package main

import (
	"fmt"
	"sync"
)

func main() {
	m := make(map[int]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mu.Lock()
			m[i] = i * 2
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	mu.Lock()
	fmt.Println("Size:", len(m))
	mu.Unlock()
}

