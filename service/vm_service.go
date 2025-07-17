package service

import (
	"context"
	"go-postgres-pagination/models"
	"go-postgres-pagination/repository"
)

type VMService interface {
	GetPaginatedVMs(ctx context.Context, page, limit int, filters map[string][]string) (*models.PaginatedVMs, error)
	SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error)
	GetVMByID(ctx context.Context, vmid string) (*models.VM, error)
}

type vmService struct {
	repo repository.VMRepository
}

func NewVMService(repo repository.VMRepository) VMService {
	return &vmService{repo}
}

func (s *vmService) GetPaginatedVMs(ctx context.Context, page, limit int, filters map[string][]string) (*models.PaginatedVMs, error) {
	data, err := s.repo.GetVMs(ctx, page, limit, filters)
	if err != nil {
		return nil, err
	}

	totalItems, err := s.repo.CountVMs(ctx, filters)
	if err != nil {
		return nil, err
	}

	totalPages := (totalItems + limit - 1) / limit

	return &models.PaginatedVMs{
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *vmService) SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error) {
	return s.repo.SearchColumnValues(ctx, column, query, limit)
}

func (s *vmService) GetVMByID(ctx context.Context, vmid string) (*models.VM, error) {
	return s.repo.GetVMByID(ctx, vmid)
}
