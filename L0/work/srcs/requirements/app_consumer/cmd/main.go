package main

import (
    "context"
    "log"
    "os"
	
    "wbtask/internal/storage"
    "wbtask/internal/kafka"
    "wbtask/internal/cache"
    "wbtask/internal/handler"
    "wbtask/internal/repository"
    "github.com/gin-gonic/gin"

    // "github.com/swaggo/gin-swagger"
    // "github.com/swaggo/files"
    // "subscription_service/docs"
)

// main — основная функция запуска сервиса
func main() {
    ctx := context.Background()

    // Получаем строку подключения к базе данных из переменных окружения
    dsn := os.Getenv("DSN")
    if dsn == "" {
        log.Fatal("Переменная окружения DSN не установлена")
    }

    log.Println("Попытка подключения к базе данных...")
    // Создаем пул подключений к PostgreSQL
    db, err := storage.NewPostgres(ctx, dsn)
    if err != nil {
        log.Fatalf("Ошибка подключения к базе данных: %v", err)
    }
    // Закрываем пул при завершении работы программы
    defer func() {
        log.Println("Закрытие подключения к базе данных")
        db.Close()
    }()
    log.Println("Подключение к базе данных успешно")

    // создаем кэш
	orderCache, err := cache.NewOrderCache(100) // лимит 100 заказов
    if err != nil {
        log.Fatalf("Не удалось создать кэш: %v", err)
    }
    // --- загружаем последние 100 заказов из БД ---
    orders, err := repository.GetLastOrders(ctx, db, 100)
    if err != nil {
        log.Fatalf("Ошибка загрузки последних заказов из БД: %v", err)
    }
    for _, o := range orders {
        orderCache.Set(o.OrderUID, o)
    }
    log.Printf("Кэш загружен последними %d заказами", len(orders))

    // Запускаем Kafka consumer в отдельной горутине
    go kafka.StartConsumer(ctx, db, orderCache)
    
    
    // Создаем роутер Gin — HTTP сервер
    router := gin.Default()
    orderHandler := handler.NewOrderHandler(db, orderCache)
    
    router.StaticFile("/", "/app/web/index.html")
    // Регистрируем маршрут (HTTP эндпоинт) и связываем его с обработчиком
    router.GET("/order/:order_uid", orderHandler.GetOrderByIDHandler)
    
    
    log.Println("Запуск сервера на порту :8081")
    // Запускаем HTTP сервер на порту 8081
    if err := router.Run(":8081"); err != nil {
        log.Fatalf("Ошибка при запуске сервера: %v", err)
    }
    // 
}

// docs.SwaggerInfo.BasePath = "/"
// // подключаем Swagger UI по пути /swagger/index.html
// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))