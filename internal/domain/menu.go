package domain

import "context"

type ServiceMenu struct {
	ID          int64   `json:"id"`
	ServiceId   int64   `json:"service_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	ImagePath   string  `json:"image_path"`
}

type Menu struct {
	ID          int64   `json:"id"`
	BranchID   int64   `json:"service_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	ImagePath   string  `json:"image_path"`
}

type MenuRepository interface {
	FindByIds(ctx context.Context, itemIds []int64) ([]Menu, error)
}