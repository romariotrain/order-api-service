package database

import (
	"context"
	"database/sql"
	"fmt"
	"order-api-service/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// OrderStore паттерн Repository
type OrderStore struct {
	db *sqlx.DB
}

func NewOrderStore(db *sqlx.DB) *OrderStore {
	return &OrderStore{db}
}

func (s *OrderStore) GetAll() ([]models.Order, error) {
	var orders []models.Order
	query := `SELECT * FROM orders`

	err := s.db.Select(&orders, query)
	// Проверяем наличие ошибки
	if err != nil {
		// Возвращаем nil (пустой слайс) и ошибку с описанием
		return nil, fmt.Errorf("ошибка получения всех задач: %w", err)
	}

	return orders, nil
}

func (s *OrderStore) Create(order *models.Order) error {
	query := `
        INSERT INTO orders (status, version, fail_reason_code)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `

	return s.db.QueryRowx(
		query,
		order.Status,
		order.Version,
		order.FailReasonCode,
	).Scan(&order.ID, &order.CreatedAt)
}

func (s *OrderStore) GetByID(ctx context.Context, id uuid.UUID) (models.Order, error) {
	const q = `
		SELECT id, status, version, fail_reason_code, fail_reason_detail, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var o models.Order
	err := s.db.GetContext(ctx, &o, q, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Order{}, models.ErrNotFound // см. ниже в models/errors.go
		}
		return models.Order{}, fmt.Errorf("get order by id: %w", err)
	}

	return o, nil
}
