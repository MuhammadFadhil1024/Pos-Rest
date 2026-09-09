package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"pos-backend/internal/domain"
	"strings"
)

type serviceRespository struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) domain.ServiceRepository {
	return &serviceRespository{db: db}
}

func (r *serviceRespository) GetByBranchId(ctx context.Context, id int64) ([]*domain.Service, error) {
	query := `SELECT s.id, s.name, s.icon 
		FROM services s 
		JOIN mm_branch_service bs ON s.id = bs.service_id
		WHERE s.is_active = 1 
		AND bs.branch_id = ?
	`
	rows , err := r.db.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Icon); err != nil {
			return  nil, err
		}
		services = append(services, &s)
	}

	return services, nil
}

func (r *serviceRespository) GetServiceMenu(ctx context.Context, branch_id int64) ([]*domain.ServiceMenu, error) {
	services, err := r.GetByBranchId(ctx, branch_id)
	if err != nil {
		return  nil, err
	}

	if len(services) == 0 {
        return []*domain.ServiceMenu{}, nil
    }

	var servicesIds []interface{}
	var placeHolders []string



	for _, service := range services {
		servicesIds =  append(servicesIds, service.ID)
		placeHolders = append(placeHolders, "?")
	}

	inClause := strings.Join(placeHolders, ",")

	// Gunakan fmt.Sprintf untuk memasukkan "?, ?, ?" ke dalam klausa IN
    query := fmt.Sprintf(`SELECT m.id as menu_id, s.id as service_id, m.name as menu_name, 
            m.price as price, m.description as description, m.image_path as image_path 
            FROM menus m 
            JOIN mm_service_menu sm ON m.id = sm.menu_id
            JOIN services s on sm.service_id = s.id
            WHERE m.branch_id = ? AND s.id IN (%s)`, inClause)
	
	args := []interface{}{branch_id}
	args = append(args, servicesIds...)

	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		return  nil, err
	}

	defer rows.Close()

	var menus []*domain.ServiceMenu

	for rows.Next() {
		var m domain.ServiceMenu
		var nullImagePath sql.NullString
		if err := rows.Scan(&m.ID, &m.ServiceId, &m.Name, &m.Price, &m.Description, &nullImagePath); err != nil {
			return nil,err
		}
		if nullImagePath.Valid {
			m.ImagePath = nullImagePath.String
		} else {
			m.ImagePath = "" // Set string kosong jika di database NULL
		}
		menus = append(menus, &m)
	}

	return  menus, nil
}