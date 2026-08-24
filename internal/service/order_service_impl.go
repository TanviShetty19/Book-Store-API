package service

import (
	"context"
	"fmt"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/dto"
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"

	"github.com/google/uuid"
)

type orderServiceImpl struct {
	orderRepo repository.OrderRepository
	bookRepo  repository.BookRepository
}

func NewOrderService(orderRepo repository.OrderRepository, bookRepo repository.BookRepository) OrderService {
	return &orderServiceImpl{
		orderRepo: orderRepo,
		bookRepo:  bookRepo,
	}
}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: order must contain at least one item", apperrors.ErrValidation)
	}

	mergedQuantities := make(map[string]int)
	var orderedBookIDs []string

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: invalid quantity %d for book %s", apperrors.ErrValidation, item.Quantity, item.BookID)
		}
		if _, exists := mergedQuantities[item.BookID]; !exists {
			orderedBookIDs = append(orderedBookIDs, item.BookID)
		}
		mergedQuantities[item.BookID] += item.Quantity
	}

	var orderItems []model.OrderItem
	var totalPrice float64

	for _, bookID := range orderedBookIDs {
		qty := mergedQuantities[bookID]
		book, err := s.bookRepo.GetByID(ctx, bookID)
		if err != nil {
			return nil, fmt.Errorf("%w: book with ID %s not found", apperrors.ErrNotFound, bookID)
		}

		if !book.CanFulfill(qty) {
			return nil, fmt.Errorf("%w: insufficient stock for '%s': requested %d, available %d", apperrors.ErrConflict, book.Title, qty, book.Stock)
		}

		unitPrice := book.Price
		lineTotal := unitPrice * float64(qty)
		totalPrice += lineTotal

		orderItems = append(orderItems, model.OrderItem{
			BookID:    bookID,
			Quantity:  qty,
			UnitPrice: unitPrice,
		})
	}

	order := &model.Order{
		ID:         uuid.New().String(),
		UserID:     userID,
		Items:      orderItems,
		TotalPrice: totalPrice,
		Status:     model.OrderStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("%w: domain validation failed: %v", apperrors.ErrValidation, err)
	}

	if err := s.orderRepo.CreateOrderWithStockDeduction(ctx, order); err != nil {
		return nil, err
	}

	return mapOrderToDTO(order), nil
}

func (s *orderServiceImpl) GetOrderByID(ctx context.Context, userID string, userRole string, orderID string) (*dto.OrderResponseDTO, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: order not found", apperrors.ErrNotFound)
	}

	if userRole != string(model.RoleAdmin) && order.UserID != userID {
		return nil, fmt.Errorf("%w: unauthorized access to order", apperrors.ErrForbidden)
	}

	return mapOrderToDTO(order), nil
}

func (s *orderServiceImpl) GetUserOrders(ctx context.Context, userID string) ([]*dto.OrderResponseDTO, error) {
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var dtos []*dto.OrderResponseDTO
	for _, order := range orders {
		dtos = append(dtos, mapOrderToDTO(order))
	}
	return dtos, nil
}

func mapOrderToDTO(order *model.Order) *dto.OrderResponseDTO {
	var itemDTOs []dto.OrderItemResponseDTO
	for _, item := range order.Items {
		itemDTOs = append(itemDTOs, dto.OrderItemResponseDTO{
			BookID:    item.BookID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return &dto.OrderResponseDTO{
		ID:         order.ID,
		UserID:     order.UserID,
		Items:      itemDTOs,
		TotalPrice: order.TotalPrice,
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
	}
}