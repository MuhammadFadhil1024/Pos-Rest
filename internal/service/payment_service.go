package service

import (
	"context"
	"pos-backend/internal/domain"
)

type paymentService struct {
	paymentRepo domain.PaymentRepository
}

func NewPaymentService(repo domain.PaymentRepository) domain.PaymentService {
	return &paymentService{paymentRepo: repo}
}

func (s *paymentService) GetServicePayments(ctx context.Context, id int64) ([]*domain.Payment, error) {
	return s.paymentRepo.GetByService(ctx, id)
}