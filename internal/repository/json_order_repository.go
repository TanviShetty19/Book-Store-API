package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

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

func (r *JsonOrderRepository) CreateOrderWithStockDeduction(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Deduct stock using the injected BookRepository
	for _, item := range order.Items {
		book, err := r.bookRepo.GetByID(ctx, item.BookID)
		if err != nil {
			return fmt.Errorf("book with ID %s not found: %w", item.BookID, err)
		}

		if !book.CanFulfill(item.Quantity) {
			return fmt.Errorf("insufficient stock for book '%s': requested %d, available %d", book.Title, item.Quantity, book.Stock)
		}

		book.Stock -= item.Quantity
		if err := r.bookRepo.Update(ctx, book); err != nil {
			return fmt.Errorf("failed to deduct stock for book '%s': %w", book.Title, err)
		}
	}

	// 2. Load and persist updated orders list
	orders, err := r.loadOrders()
	if err != nil {
		return fmt.Errorf("failed to read orders storage: %w", err)
	}

	orders = append(orders, *order)
	return r.saveOrders(orders)
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