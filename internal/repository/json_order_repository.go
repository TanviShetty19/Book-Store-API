package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"bookstore-api/internal/model"
)

type JsonOrderRepository struct {
	mu            sync.RWMutex
	orderFilePath string
	bookRepo      BookRepository
}

func NewJsonOrderRepository(orderFilePath string, bookRepo BookRepository) (*JsonOrderRepository, error) {
	repo := &JsonOrderRepository{
		orderFilePath: orderFilePath,
		bookRepo:      bookRepo,
	}
	if err := repo.initStorage(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JsonOrderRepository) initStorage() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := os.Stat(r.orderFilePath); os.IsNotExist(err) {
		data, _ := json.MarshalIndent([]model.Order{}, "", "  ")
		if err := os.WriteFile(r.orderFilePath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (r *JsonOrderRepository) loadOrders() ([]model.Order, error) {
	ordersData, err := os.ReadFile(r.orderFilePath)
	if err != nil {
		return nil, err
	}
	var orders []model.Order
	if err := json.Unmarshal(ordersData, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *JsonOrderRepository) saveOrders(orders []model.Order) error {
	updatedOrdersData, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated orders: %w", err)
	}
	if err := os.WriteFile(r.orderFilePath, updatedOrdersData, 0644); err != nil {
		return fmt.Errorf("failed to write updated orders: %w", err)
	}
	return nil
}

// CreateDraftOrder persists a new order in PENDING (DRAFT) status.
// Cart model: no stock reservation/deduction happens here. Multiple
// draft orders may reference the same last unit of stock; contention
// is resolved later, atomically, in ConfirmOrder.
func (r *JsonOrderRepository) CreateDraftOrder(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return fmt.Errorf("failed to read orders storage: %w", err)
	}

	orders = append(orders, *order)
	return r.saveOrders(orders)
}

// ConfirmOrder performs the authoritative, atomic transition from PENDING
// to CONFIRMED. The mock payment delay is expected to happen in the
// service layer BEFORE this is called, so the lock here only guards the
// final re-check + stock deduction + persist sequence.
func (r *JsonOrderRepository) ConfirmOrder(ctx context.Context, orderID string) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to read orders storage: %w", err)
	}

	idx := -1
	for i, o := range orders {
		if o.ID == orderID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, errors.New("order not found")
	}

	if orders[idx].Status != model.OrderStatusPending {
		return nil, fmt.Errorf("order is not pending (current status: %s)", orders[idx].Status)
	}

	// Pass 1: validate stock availability for every item before mutating anything.
	books := make(map[string]*model.Book, len(orders[idx].Items))
	for _, item := range orders[idx].Items {
		book, err := r.bookRepo.GetByID(ctx, item.BookID)
		if err != nil {
			return nil, fmt.Errorf("book with ID %s not found: %w", item.BookID, err)
		}
		if !book.CanFulfill(item.Quantity) {
			return nil, fmt.Errorf("insufficient stock for '%s': requested %d, available %d", book.Title, item.Quantity, book.Stock)
		}
		books[item.BookID] = book
	}

	// Pass 2: all checks passed, apply deductions.
	for _, item := range orders[idx].Items {
		book := books[item.BookID]
		book.Stock -= item.Quantity
		if err := r.bookRepo.Update(ctx, book); err != nil {
			return nil, fmt.Errorf("failed to deduct stock for book '%s': %w", book.Title, err)
		}
	}

	orders[idx].Status = model.OrderStatusCompleted
	orders[idx].UpdatedAt = time.Now()

	if err := r.saveOrders(orders); err != nil {
		return nil, err
	}

	confirmed := orders[idx]
	return &confirmed, nil
}

// CancelOrder transitions a PENDING order to CANCELLED. No stock changes
// are needed since the cart model never reserved stock at DRAFT time.
func (r *JsonOrderRepository) CancelOrder(ctx context.Context, orderID string) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to read orders storage: %w", err)
	}

	idx := -1
	for i, o := range orders {
		if o.ID == orderID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, errors.New("order not found")
	}

	if orders[idx].Status != model.OrderStatusPending {
		return nil, fmt.Errorf("order is not pending (current status: %s)", orders[idx].Status)
	}

	orders[idx].Status = model.OrderStatusCancelled
	orders[idx].UpdatedAt = time.Now()

	if err := r.saveOrders(orders); err != nil {
		return nil, err
	}

	cancelled := orders[idx]
	return &cancelled, nil
}

func (r *JsonOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, err := r.loadOrders()
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		if order.ID == id {
			return &order, nil
		}
	}
	return nil, errors.New("order not found")
}

func (r *JsonOrderRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, err := r.loadOrders()
	if err != nil {
		return nil, err
	}

	var userOrders []*model.Order
	for i := range orders {
		if orders[i].UserID == userID {
			userOrders = append(userOrders, &orders[i])
		}
	}
	return userOrders, nil
}
