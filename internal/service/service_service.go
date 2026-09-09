package service

import (
	"context"
	"pos-backend/internal/domain"
)

type serviceTypeService struct {
	serviceRepo domain.ServiceRepository
	servicePaymentRepo domain.PaymentRepository
}

func NewServiceTypeService(repo domain.ServiceRepository, paymentRepo domain.PaymentRepository) domain.ServiceTypeService {
	return &serviceTypeService{
		serviceRepo:        repo,
		servicePaymentRepo: paymentRepo,
	}
}

func (s *serviceTypeService) GetBranchServices(ctx context.Context, id int64) ([]*domain.Service, error) {
	// return s.serviceRepo.GetByBranchId(ctx, id)

	services, err := s.serviceRepo.GetByBranchId(ctx, id)

	if err != nil {
		return nil, err
	}

	payments, err := s.servicePaymentRepo.GetByService(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a map to associate service IDs with their corresponding payments
	paymentMap := make(map[int64][]*domain.Payment)
	for _, payment := range payments {
		paymentMap[payment.ServiceID] = append(paymentMap[payment.ServiceID], payment)
	}

	// Assign payments to the corresponding services
	for _, service := range services {
		servicePayments, ok := paymentMap[service.ID]
		if ok {
			service.Payments = servicePayments
		} else {
			service.Payments = []*domain.Payment{}
		}
	}


	return services, nil
}

func (s *serviceTypeService) GetBranchServiceMenu(ctx context.Context, branch_id int64) ([]*domain.ServiceMenu, error) {
	return s.serviceRepo.GetServiceMenu(ctx, branch_id)
}