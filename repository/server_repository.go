package repository

import (
	"context"
	"database/sql"
	"fmt"

	"strings"

	"go-postgres-pagination/models"
)

type ServerRepository interface {
	GetServers(ctx context.Context, page, limit int, filters map[string][]string) ([]models.Server, error)
	CountServers(ctx context.Context, filters map[string][]string) (int, error)
	SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error)
	GetServerByID(ctx context.Context, id int) (*models.Server, error)
}

func (r *serverRepo) SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error) {
	var sqlQuery string
	var args []interface{}

	switch column {
	case "customer_name", "name":
		// Text alanları için ILIKE kullan
		sqlQuery = fmt.Sprintf(`
			SELECT DISTINCT %s 
			FROM servers 
			WHERE %s ILIKE $1 
			ORDER BY %s 
			LIMIT $2
		`, column, column, column)
		args = []interface{}{"%" + query + "%", limit}
	case "ip_address", "cpu", "memory", "status":
		// Exact match alanları için = kullan
		sqlQuery = fmt.Sprintf(`
			SELECT DISTINCT %s 
			FROM servers 
			WHERE %s = $1 
			ORDER BY %s 
			LIMIT $2
		`, column, column, column)
		args = []interface{}{query, limit}
	default:
		return nil, fmt.Errorf("unsupported column: %s", column)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, value)
	}

	return suggestions, nil
}

func NewServerRepository(db *sql.DB) ServerRepository {
	return &serverRepo{db}
}

type serverRepo struct {
	db *sql.DB
}

func (r *serverRepo) buildFilterSQL(base string, filters map[string][]string, args *[]interface{}) string {
	i := len(*args) + 1

	for column, values := range filters {
		switch column {
		case "customer_name", "name":
			if len(values) > 1 {
				base += " AND ("
				for j, v := range values {
					if j > 0 {
						base += " OR "
					}
					base += fmt.Sprintf("%s ILIKE $%d", column, i)
					*args = append(*args, "%"+v+"%")
					i++
				}
				base += ")"
			} else {
				base += fmt.Sprintf(" AND %s ILIKE $%d", column, i)
				*args = append(*args, "%"+values[0]+"%")
				i++
			}
		case "ip_address", "cpu", "memory", "status":
			if len(values) > 1 {
				base += " AND ("
				for j, v := range values {
					if j > 0 {
						base += " OR "
					}
					base += fmt.Sprintf("%s = $%d", column, i)
					*args = append(*args, v)
					i++
				}
				base += ")"
			} else {
				base += fmt.Sprintf(" AND %s = $%d", column, i)
				*args = append(*args, values[0])
				i++
			}
		default:
			continue
		}
	}

	return base
}

func joinWithOr(conditions []string) string {
	return strings.Join(conditions, " OR ")
}

func (r *serverRepo) GetServers(ctx context.Context, page, limit int, filters map[string][]string) ([]models.Server, error) {
	offset := (page - 1) * limit
	args := []interface{}{}

	base := "SELECT id, customer_name, name, ip_address, cpu, memory, status FROM servers WHERE 1=1"
	base = r.buildFilterSQL(base, filters, &args)

	args = append(args, limit, offset)
	query := fmt.Sprintf("%s ORDER BY id LIMIT $%d OFFSET $%d", base, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var s models.Server
		if err := rows.Scan(&s.ID, &s.CustomerName, &s.Name, &s.IPAddress, &s.CPU, &s.Memory, &s.Status); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, nil
}

func (r *serverRepo) CountServers(ctx context.Context, filters map[string][]string) (int, error) {
	args := []interface{}{}
	base := "SELECT COUNT(*) FROM servers WHERE 1=1"
	base = r.buildFilterSQL(base, filters, &args)

	var count int
	err := r.db.QueryRowContext(ctx, base, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *serverRepo) GetServerByID(ctx context.Context, id int) (*models.Server, error) {
	query := "SELECT id, customer_name, name, ip_address, cpu, memory, status FROM servers WHERE id = $1"

	var server models.Server
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&server.ID,
		&server.CustomerName,
		&server.Name,
		&server.IPAddress,
		&server.CPU,
		&server.Memory,
		&server.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Server bulunamadı
		}
		return nil, err
	}

	return &server, nil
}
