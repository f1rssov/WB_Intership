package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== 1. Остановка по условию (флаг) ===")
	stopByFlag()

	fmt.Println("\n=== 2. Через канал уведомления ===")
	stopByChannel()

	fmt.Println("\n=== 3. Через context.WithCancel ===")
	stopByContext()

	
	fmt.Println("\n=== 5. Завершение через return ===")
	stopByReturn()
	
	fmt.Println("\n=== 6. Принудительное завершение программы (os.Exit) ===")
	stopByOsExit()
	
	fmt.Println("\n=== 7. Завершение через panic/recover ===")
	stopByPanicRecover()
	
	fmt.Println("\n=== 4. Через runtime.Goexit() ===")
	stopByGoexit()
}

// 1. Остановка по условию
func stopByFlag() {
	var stopFlag atomic.Bool

	go func() {
		for !stopFlag.Load() {
			fmt.Println("Работаю...")
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println("Горутина завершена (по флагу)")
	}()

	time.Sleep(2 * time.Second)
	stopFlag.Store(true)
	time.Sleep(500 * time.Millisecond)
}

// 2. Остановка через канал уведомления
func stopByChannel() {
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				fmt.Println("Горутина завершена (через канал)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	time.Sleep(2 * time.Second)
	close(stop)
	time.Sleep(500 * time.Millisecond)
}

// 3. Через контекст
func stopByContext() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Горутина завершена (через context)")
				return
			default:
				fmt.Println("Работаю...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(500 * time.Millisecond)
}

// 4. Через runtime.Goexit()
func stopByGoexit() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fmt.Println("Горутина запущена")
		time.Sleep(time.Second)
		fmt.Println("Завершаюсь через runtime.Goexit()")
		runtime.Goexit()
		fmt.Println("Эта строка не будет выполнена")
	}()

	wg.Wait()
}

// 5. Через return
func stopByReturn() {
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 3; i++ {
			fmt.Println("Работаю...")
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println("Горутина завершена (return)")
		return
	}()

	<-done
}

// 6. Принудительное завершение программы (os.Exit)
func stopByOsExit() {
	go func() {
		fmt.Println("Работаю...")
		time.Sleep(1 * time.Second)
		fmt.Println("Завершаю всю программу через os.Exit(0)")
		os.Exit(0)
	}()

	time.Sleep(3 * time.Second) // никогда не выполнится
}

// 7. Завершение через panic/recover
func stopByPanicRecover() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Поймано panic:", r)
			}
		}()
		for i := 0; i < 3; i++ {
			if i == 2 {
				panic("остановка горутины через panic")
			}
			fmt.Println("Работаю...")
			time.Sleep(500 * time.Millisecond)
		}
	}()

	wg.Wait()
}
