package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"pos-backend/internal/domain"
	"strings"
)

type menuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) domain.MenuRepository {
	return &menuRepository{db:db}
}

func (r *menuRepository) FindByIds(ctx context.Context, menuIDs []int64) ([]domain.Menu, error)  {

	if len(menuIDs) == 0 {
		return []domain.Menu{}, nil
	}

	placeholders := make([]string, len(menuIDs))
	args := make([]any, len(menuIDs))

	for i, id := range menuIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, branch_id, name, price, description, image_path FROM menus WHERE id IN (%s)	
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return  nil, err
	}
	defer rows.Close()
	menus := make([]domain.Menu, 0, len(menuIDs))

	for rows.Next() {
		var menu domain.Menu
		var nullDescription, nullImagePath sql.NullString

		err := rows.Scan(
			&menu.ID,
			&menu.BranchID,
			&menu.Name,
			&menu.Price,
			&nullDescription,
			&nullImagePath,
		)

		if err != nil {
			return nil, err
		}

		if nullDescription.Valid {
			menu.Description = nullDescription.String
		}
		if nullImagePath.Valid {
			menu.ImagePath = nullImagePath.String
		}

		menus = append(menus, menu)
	}
	if err := rows.Err(); err != nil {
		return nil, err		
	}

	return menus, nil
}
