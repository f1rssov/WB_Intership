package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "worker-pool",
		Usage: "A program with workers that process data from a channel",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "workers",
				Aliases:  []string{"w"},
				Required: true,
				Usage:    "Number of worker goroutines",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			workerCount := c.Int("workers")
			if workerCount <= 0 {
				return fmt.Errorf("number of workers must be positive, got %d", workerCount)
			}

			dataChan := make(chan string, 100)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			for i := 0; i < workerCount; i++ {
				wg.Add(1)
				go worker(ctx, i+1, dataChan, &wg)
			}

			wg.Add(1)
			go dataProducer(ctx, dataChan, &wg)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			<-sigChan
			fmt.Println("\nReceived shutdown signal, stopping...")
			cancel()
			close(dataChan)

			wg.Wait()
			fmt.Println("All workers stopped")

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// dataProducer постоянно записывает данные в канал
func dataProducer(ctx context.Context, dataChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	counter := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			counter++
			data := fmt.Sprintf("Data item #%d", counter)

			select {
			case dataChan <- data:
				// Данные успешно отправлены
			case <-ctx.Done():
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// worker читает данные из канала и выводит их в stdout
func worker(ctx context.Context, id int, dataChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				fmt.Printf("Worker %d: channel closed, stopping\n", id)
				return
			}
			fmt.Printf("Worker %d: %s\n", id, data)

		case <-ctx.Done():
			fmt.Printf("Worker %d: received shutdown signal, stopping\n", id)
			return
		}
	}
}
