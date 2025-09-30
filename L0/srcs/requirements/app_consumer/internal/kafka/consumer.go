package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"wbtask/internal/cache"
	"wbtask/internal/model"
	"wbtask/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"fmt"
)

var v = validator.New()

func StartConsumer(ctx context.Context, db *pgxpool.Pool, orderCache *cache.OrderCache) {
	waitForKafka()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "orders",
		GroupID: fmt.Sprintf("orders-consumer-group-%d", time.Now().Unix()),
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	log.Println("Kafka consumer запущен и подписан на топик orders")

	for {
		if err := handleMessage(ctx, r, db, orderCache); err != nil {
			log.Println("Ошибка обработки сообщения:", err)
		}
	}
}

func waitForKafka() {
	for {
		conn, err := kafka.Dial("tcp", "kafka:9092")
		if err != nil {
			log.Println("Kafka не готова, повтор через 3 секунды...")
			time.Sleep(3 * time.Second)
			continue
		}
		conn.Close()
		log.Println("Kafka готова, запускаем consumer")
		break
	}
}

func handleMessage(ctx context.Context, r *kafka.Reader, db *pgxpool.Pool, orderCache *cache.OrderCache) error {
	log.Printf("Ожидание сообщения...")
	log.Printf("...")
	m, err := r.ReadMessage(ctx)
	if err != nil {
		return err
	}

	var order model.Order
	if err := json.Unmarshal(m.Value, &order); err != nil {
		return err
	}

	if err := v.Struct(order); err != nil {
		return err
	}

	if err := repository.InsertOrder(ctx, db, &order); err != nil {
		return err
	}

	updatedOrder, err := repository.GetOrderByUID(ctx, db, order.OrderUID)
	if err != nil {
		return err
	}

	orderCache.Set(order.OrderUID, updatedOrder)
	log.Println("Заказ успешно сохранён:", order.OrderUID)

	return r.CommitMessages(ctx, m)
}
