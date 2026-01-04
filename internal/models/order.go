package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "NEW"
	OrderStatusReserved  OrderStatus = "RESERVED"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type Order struct {
	ID               uuid.UUID   `json:"id" db:"id"`
	Status           OrderStatus `json:"status" db:"status"`
	Version          int64       `json:"version" db:"version"`
	FailReasonCode   *string     `json:"fail_reason_code,omitempty" db:"fail_reason_code"`
	FailReasonDetail *string     `json:"fail_reason_detail,omitempty" db:"fail_reason_detail"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" db:"updated_at"`
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	switch s {
	case OrderStatusNew:
		return next == OrderStatusReserved || next == OrderStatusFailed
	case OrderStatusReserved:
		return next == OrderStatusConfirmed || next == OrderStatusFailed
	case OrderStatusConfirmed, OrderStatusFailed:
		return false
	default:
		return false
	}
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusNew,
		OrderStatusReserved,
		OrderStatusConfirmed,
		OrderStatusFailed:
		return true
	default:
		return false
	}
}

func (o *Order) TransitionTo(next OrderStatus, now time.Time) error {

	if o.Status == OrderStatusConfirmed || o.Status == OrderStatusFailed {
		return ErrTerminalState
	}

	if o.Status == next {
		return nil
	}

	if !next.IsValid() || !o.Status.IsValid() {
		return ErrInvalidStatus
	}

	if !o.Status.CanTransitionTo(next) {
		return ErrInvalidTransition
	}

	o.Status = next
	o.Version++
	o.UpdatedAt = now

	return nil
}
