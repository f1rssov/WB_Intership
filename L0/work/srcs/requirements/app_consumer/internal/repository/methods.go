package repository

import (
    "context"
    "fmt"
    "wbtask/internal/model"
    "github.com/jackc/pgx/v5/pgxpool"
)


// InsertOrder вставляет заказ и связанные сущности в БД в одной транзакции.
// Если заказ уже есть, дубликаты не создаются.
func InsertOrder(ctx context.Context, db *pgxpool.Pool, order *model.Order) error {
    // Начинаем транзакцию
    tx, err := db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("не удалось начать транзакцию: %w", err)
    }
    defer tx.Rollback(ctx) // откат в случае ошибки или panic

    // Вставка в orders
    _, err = tx.Exec(ctx, `
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id,
                            delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        ON CONFLICT (order_uid) DO NOTHING
    `, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
        order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
    if err != nil {
        return fmt.Errorf("ошибка вставки заказа: %w", err)
    }

    // Вставка в delivery
    _, err = tx.Exec(ctx, `
        INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (order_uid) DO UPDATE
        SET name = EXCLUDED.name, phone = EXCLUDED.phone, zip = EXCLUDED.zip,
            city = EXCLUDED.city, address = EXCLUDED.address,
            region = EXCLUDED.region, email = EXCLUDED.email
    `, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
        order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
    if err != nil {
        return fmt.Errorf("ошибка вставки delivery: %w", err)
    }

    // Вставка в payment
    _, err = tx.Exec(ctx, `
        INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount,
                             payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        ON CONFLICT (order_uid) DO UPDATE
        SET transaction=EXCLUDED.transaction, request_id=EXCLUDED.request_id, 
            currency=EXCLUDED.currency, provider=EXCLUDED.provider, amount=EXCLUDED.amount,
            payment_dt=EXCLUDED.payment_dt, bank=EXCLUDED.bank, delivery_cost=EXCLUDED.delivery_cost,
            goods_total=EXCLUDED.goods_total, custom_fee=EXCLUDED.custom_fee
    `, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
        order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
        order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
    if err != nil {
        return fmt.Errorf("ошибка вставки payment: %w", err)
    }

    // Вставка items
    for _, item := range order.Items {
        _, err = tx.Exec(ctx, `
            INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
        `, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
            item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status)
        if err != nil {
            return fmt.Errorf("ошибка вставки item: %w", err)
        }
    }

    // Коммит транзакции
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
    }

    return nil
}

// GetOrderByUID получает заказ со всеми вложенными сущностями
func GetOrderByUID(ctx context.Context, db *pgxpool.Pool, orderUID string) (*model.Order, error) {
    var order model.Order

    // 1. Основные данные + delivery + payment (через JOIN)
    query := `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
            o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
            
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
            
            p.transaction, p.request_id, p.currency, p.provider, p.amount,
            p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        JOIN delivery d ON o.order_uid = d.order_uid
        JOIN payment p ON o.order_uid = p.order_uid
        WHERE o.order_uid = $1
    `

    row := db.QueryRow(ctx, query, orderUID)

    err := row.Scan(
        &order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
        &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
        &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,

        &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
        &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
        &order.Delivery.Email,

        &order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
        &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT,
        &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
        &order.Payment.CustomFee,
    )
    if err != nil {
        return nil, fmt.Errorf("ошибка получения заказа: %w", err)
    }

    // 2. Товары
    itemsQuery := `
        SELECT chrt_id, track_number, price, rid, name, sale, size,
               total_price, nm_id, brand, status
        FROM items
        WHERE order_uid = $1
    `
    rows, err := db.Query(ctx, itemsQuery, orderUID)
    if err != nil {
        return nil, fmt.Errorf("ошибка получения items: %w", err)
    }
    defer rows.Close()

    var items []model.Item
    for rows.Next() {
        var item model.Item
        err := rows.Scan(
            &item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
            &item.Sale, &item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
        )
        if err != nil {
            return nil, fmt.Errorf("ошибка чтения item: %w", err)
        }
        items = append(items, item)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("ошибка при обходе items: %w", err)
    }

    order.Items = items

    return &order, nil
}

// Получает последние N заказов из БД
func GetLastOrders(ctx context.Context, db *pgxpool.Pool, limit int) ([]*model.Order, error) {
	query := `
	SELECT o.order_uid
	FROM orders o
	ORDER BY o.date_created DESC
	LIMIT $1
	`
	rows, err := db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса последних заказов: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("ошибка сканирования order_uid: %w", err)
		}

		order, err := GetOrderByUID(ctx, db, orderUID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения заказа %s: %w", orderUID, err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
