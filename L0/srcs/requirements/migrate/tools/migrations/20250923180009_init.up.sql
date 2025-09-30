-- Основная таблица заказов
CREATE TABLE orders (
    order_uid VARCHAR(64) PRIMARY KEY,
    track_number VARCHAR(64) NOT NULL,
    entry VARCHAR(16),
    locale VARCHAR(8),
    internal_signature VARCHAR(128),
    customer_id VARCHAR(64),
    delivery_service VARCHAR(64),
    shardkey VARCHAR(8),
    sm_id INT,
    date_created TIMESTAMP NOT NULL,
    oof_shard VARCHAR(8)
);

-- Доставка (один к одному с заказом)
CREATE TABLE delivery (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128),
    phone VARCHAR(32),
    zip VARCHAR(16),
    city VARCHAR(64),
    address VARCHAR(128),
    region VARCHAR(64),
    email VARCHAR(128),
    CONSTRAINT fk_delivery_order FOREIGN KEY (order_uid)
        REFERENCES orders(order_uid) ON DELETE CASCADE
);

-- Оплата (один к одному с заказом)
CREATE TABLE payment (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(64) NOT NULL UNIQUE,
    transaction VARCHAR(64),
    request_id VARCHAR(64),
    currency VARCHAR(8),
    provider VARCHAR(64),
    amount INT,
    payment_dt BIGINT,
    bank VARCHAR(64),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT,
    CONSTRAINT fk_payment_order FOREIGN KEY (order_uid)
        REFERENCES orders(order_uid) ON DELETE CASCADE
);

-- Товары (многие к одному: много items на один заказ)
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(64) NOT NULL,
    chrt_id BIGINT,
    track_number VARCHAR(64),
    price INT,
    rid VARCHAR(64),
    name VARCHAR(128),
    sale INT,
    size VARCHAR(16),
    total_price INT,
    nm_id BIGINT,
    brand VARCHAR(64),
    status INT,
    CONSTRAINT fk_items_order FOREIGN KEY (order_uid)
        REFERENCES orders(order_uid) ON DELETE CASCADE
);
