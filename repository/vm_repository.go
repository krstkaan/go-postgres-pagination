package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-postgres-pagination/models"
)

type VMRepository interface {
	GetVMs(ctx context.Context, page, limit int, filters map[string][]string) ([]models.VM, error)
	CountVMs(ctx context.Context, filters map[string][]string) (int, error)
	SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error)
	GetVMByID(ctx context.Context, vmid string) (*models.VM, error)
}

type vmRepo struct {
	db *sql.DB
}

func NewVMRepository(db *sql.DB) VMRepository {
	return &vmRepo{db}
}

func (r *vmRepo) buildFilterSQL(base string, filters map[string][]string, args *[]interface{}) string {
	i := len(*args) + 1

	for column, values := range filters {
		switch column {
		case "name", "node", "cluster", "datacenter", "guestos", "hypervisor_agent":
			// Text fields with ILIKE
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
		case "vmid", "hypervisor", "status":
			// ILIKE search fields
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
		case "tags":
			// Array field search
			if len(values) > 1 {
				base += " AND ("
				for j, v := range values {
					if j > 0 {
						base += " OR "
					}
					base += fmt.Sprintf("$%d = ANY(tags)", i)
					*args = append(*args, v)
					i++
				}
				base += ")"
			} else {
				base += fmt.Sprintf(" AND $%d = ANY(tags)", i)
				*args = append(*args, values[0])
				i++
			}
		case "ip":
			// IP array field search
			if len(values) > 1 {
				base += " AND ("
				for j, v := range values {
					if j > 0 {
						base += " OR "
					}
					base += fmt.Sprintf("$%d::inet = ANY(ip)", i)
					*args = append(*args, v)
					i++
				}
				base += ")"
			} else {
				base += fmt.Sprintf(" AND $%d::inet = ANY(ip)", i)
				*args = append(*args, values[0])
				i++
			}
		default:
			continue
		}
	}

	return base
}

func (r *vmRepo) GetVMs(ctx context.Context, page, limit int, filters map[string][]string) ([]models.VM, error) {
	offset := (page - 1) * limit
	args := []interface{}{}

	base := `SELECT vmid, hypervisor, name, status, cpu, mem, disk, ip, tags, node, cluster, 
			 datacenter, guestos, hypervisor_agent, raw, updated_at, deleted_at 
			 FROM vms WHERE deleted_at IS NULL`

	base = r.buildFilterSQL(base, filters, &args)

	args = append(args, limit, offset)
	query := fmt.Sprintf("%s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", base, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []models.VM
	for rows.Next() {
		var vm models.VM
		if err := rows.Scan(
			&vm.VMId,
			&vm.Hypervisor,
			&vm.Name,
			&vm.Status,
			&vm.CPU,
			&vm.Memory,
			&vm.Disk,
			&vm.IPAddresses,
			&vm.Tags,
			&vm.Node,
			&vm.Cluster,
			&vm.Datacenter,
			&vm.GuestOS,
			&vm.HypervisorAgent,
			&vm.Raw,
			&vm.UpdatedAt,
			&vm.DeletedAt,
		); err != nil {
			return nil, err
		}
		vms = append(vms, vm)
	}
	return vms, nil
}

func (r *vmRepo) CountVMs(ctx context.Context, filters map[string][]string) (int, error) {
	args := []interface{}{}
	base := "SELECT COUNT(*) FROM vms WHERE deleted_at IS NULL"
	base = r.buildFilterSQL(base, filters, &args)

	var count int
	err := r.db.QueryRowContext(ctx, base, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *vmRepo) SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error) {
	var sqlQuery string
	var args []interface{}

	switch column {
	case "name", "node", "cluster", "datacenter", "guestos", "hypervisor_agent":
		sqlQuery = fmt.Sprintf(`
			SELECT DISTINCT %s 
			FROM vms 
			WHERE deleted_at IS NULL AND %s ILIKE $1 
			ORDER BY %s 
			LIMIT $2
		`, column, column, column)
		args = []interface{}{"%" + query + "%", limit}
	case "vmid", "hypervisor", "status":
		sqlQuery = fmt.Sprintf(`
			SELECT DISTINCT %s 
			FROM vms 
			WHERE deleted_at IS NULL AND %s ILIKE $1 
			ORDER BY %s 
			LIMIT $2
		`, column, column, column)
		args = []interface{}{"%" + query + "%", limit}
	case "tags":
		sqlQuery = `
			SELECT DISTINCT unnest(tags) as tag 
			FROM vms 
			WHERE deleted_at IS NULL AND $1 = ANY(tags)
			ORDER BY tag 
			LIMIT $2
		`
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

func (r *vmRepo) GetVMByID(ctx context.Context, vmid string) (*models.VM, error) {
	query := `SELECT vmid, hypervisor, name, status, cpu, mem, disk, ip, tags, node, cluster, 
			  datacenter, guestos, hypervisor_agent, raw, updated_at, deleted_at 
			  FROM vms WHERE vmid = $1 AND deleted_at IS NULL`

	var vm models.VM
	err := r.db.QueryRowContext(ctx, query, vmid).Scan(
		&vm.VMId,
		&vm.Hypervisor,
		&vm.Name,
		&vm.Status,
		&vm.CPU,
		&vm.Memory,
		&vm.Disk,
		&vm.IPAddresses,
		&vm.Tags,
		&vm.Node,
		&vm.Cluster,
		&vm.Datacenter,
		&vm.GuestOS,
		&vm.HypervisorAgent,
		&vm.Raw,
		&vm.UpdatedAt,
		&vm.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // VM not found
		}
		return nil, err
	}

	return &vm, nil
}
