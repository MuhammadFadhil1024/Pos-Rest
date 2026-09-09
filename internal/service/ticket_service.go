package service

import (
	"context"
	"errors"
	"log"
	"pos-backend/internal/domain"
	"pos-backend/pkg/utils"
)

type ticketService struct {
	ticketRepo domain.TicketRepository
	menuRepo   domain.MenuRepository
}

func NewTicketService(repoTicket domain.TicketRepository, repoMenu domain.MenuRepository) domain.TicketService {
	return &ticketService{
		ticketRepo: repoTicket,
		menuRepo:   repoMenu,
	}
}

func (s *ticketService) CreateTicket(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error) {
	if branchId == 0 || userId == 0 || serviceId == 0 {
		return nil, errors.New("invalid branch, user, or service id")
	}
	if len(items) == 0 {
		return nil, errors.New("ticket must contain at least one item")
	}

	menuIDs := make([]int64, 0, len(items))
	for _, reqMenu := range items {
		menuIDs = append(menuIDs, reqMenu.MenuId)
	}

	menus, err := s.menuRepo.FindByIds(ctx, menuIDs)
	if err != nil {
		return nil, err
	}

	menuPrices := make(map[int64]float64)
	
	for _, menu := range menus {
		menuPrices[menu.ID] = menu.Price
	}

	// Calculate Total Amount
	calculateTotalAmount := 0.0
	for _, item := range items {
		price := menuPrices[item.MenuId]
		calculateTotalAmount += price * float64(item.Quantity)
	}

	if payments == nil || len(payments) == 0 {
		return nil, errors.New("ticket must contain at least one payment")
	}

	totalPayment := 0.0
	for _, payment := range payments {
		totalPayment += payment.Amount
	}

	if totalPayment < calculateTotalAmount {
		return nil, errors.New("total payment is less than total amount")
	}

	// Prepare data for TicketItemInsertDb
	prepareDataMenu := make(map[int64]domain.TicketItemInsertDb)
	for _, item := range items {
		price, exists := menuPrices[item.MenuId]
		if !exists {
			return nil, errors.New("menu item not found")
		}
		subtotal := price * float64(item.Quantity)
		prepareDataMenu[item.MenuId] = domain.TicketItemInsertDb{
			MenuId:   item.MenuId,
			Quantity: item.Quantity,
			Price:    price,
			Subtotal: subtotal,
		}
	}

	// prepare data for ticket payment insert db
	prepareDataPayment := make([]domain.TicketPaymentInsertDb, 0, len(payments))
	for _, payment := range payments {
		changeAmount := payment.Amount - calculateTotalAmount
		if changeAmount < 0 {
			changeAmount = 0
		}
		prepareDataPayment = append(prepareDataPayment, domain.TicketPaymentInsertDb{
			PaymentTypeId: payment.PaymentTypeId,
			Amount:          payment.Amount,
			ChangeAmount:    changeAmount,
		})
	}

	var itemsInsertDb []domain.TicketItemInsertDb
	for _, item := range prepareDataMenu {
		itemsInsertDb = append(itemsInsertDb, item)
	}

	log.Printf("Calculated total amount: %f", calculateTotalAmount)
	ticketNumber := utils.GenerateTicketNumber(int(branchId))
	var status = "PENDING"
	if serviceId == 1 {
		status = "PAID"
	}

	ticket := &domain.Ticket{
		BranchId:     branchId,
		UserId:       userId,
		ServiceId:    serviceId,
		TicketNumber: ticketNumber,
		TotalAmount:  float64(calculateTotalAmount),
		Status:       status,
	}

	if err := s.ticketRepo.CreateFullTicketWithTx(ctx, ticket, itemsInsertDb, prepareDataPayment); err != nil {
		return nil, err
	}

	return ticket, nil
}
