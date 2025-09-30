package handler

import (
	"context"
	"net/http"
	"wbtask/internal/repository" 
	"wbtask/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderHandler struct {
	db *pgxpool.Pool
	cache *cache.OrderCache
}

// Конструктор
func NewOrderHandler(db *pgxpool.Pool, cache *cache.OrderCache) *OrderHandler {
	return &OrderHandler{db: db, cache: cache}
}

// GetOrderByIDHandler — обертка для твоей функции GetOrderByUID
func (h *OrderHandler) GetOrderByIDHandler(c *gin.Context) {
	orderUID := c.Param("order_uid")
	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_uid обязателен"})
		return
	}

	// Сначала проверяем кэш
	if order, ok := h.cache.Get(orderUID); ok {
		c.IndentedJSON(http.StatusOK, order)
		return
	}

	// Если нет в кэше, получаем из БД
	order, err := repository.GetOrderByUID(context.Background(), h.db, orderUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем в кэш
	h.cache.Set(orderUID, order)

	c.IndentedJSON(http.StatusOK, order)
}