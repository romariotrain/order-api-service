package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"order-api-service/internal/models"
	"strings"

	"github.com/google/uuid"
)

// OrderRepository — минимальный контракт, который нужен хендлеру.
// Так хендлер не зависит от конкретной реализации (sqlx/pgx/моки).
type OrderRepository interface {
	GetAll() ([]models.Order, error)
	Create(order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (models.Order, error)
}

type Handler struct {
	orders OrderRepository
}

func NewHandler(orders OrderRepository) *Handler {
	return &Handler{orders: orders}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GetAllOrders — GET /orders
func (h *Handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orders.GetAll()
	if err != nil {
		log.Printf("GetAllOrders: repo.GetAll error: %v", err)

		writeError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	// Можно возвращать массив напрямую
	writeJSON(w, http.StatusOK, orders)
}

// CreateOrder — POST /orders
// Сейчас создаём “пустой” заказ через доменный конструктор.
// Позже добавишь парсинг тела запроса и заполнение бизнес-полей.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	order := models.NewOrder()

	if err := h.orders.Create(&order); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	// минимально: процесс жив, HTTP отвечает
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	order, err := h.orders.GetByID(r.Context(), id)
	if err != nil {
		if err == models.ErrNotFound {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch order")
		return
	}

	writeJSON(w, http.StatusOK, order)
}
