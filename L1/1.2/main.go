package main

import (
	"fmt"
	"sync"
)

func main() {
	numbers := []int{2, 4, 6, 8, 10}
	var wg sync.WaitGroup

	for _, n := range numbers {
		wg.Add(1)

		// Запускаем горутину для каждого числа
		go func(num int) {
			defer wg.Done()
			square := num * num
			fmt.Printf("Квадрат числа %d = %d\n", num, square)
		}(n)
	}

	wg.Wait() // Ждем завершения всех горутин
}