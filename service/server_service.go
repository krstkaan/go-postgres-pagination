package service

import (
	"context"
	"go-postgres-pagination/models"
	"go-postgres-pagination/repository"
)

type ServerService interface {
	GetPaginatedServers(ctx context.Context, page, limit int, filters map[string][]string) (*models.PaginatedServers, error)
	SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error)
	GetServerByID(ctx context.Context, id int) (*models.Server, error)
}

type serverService struct {
	repo repository.ServerRepository
}

func NewServerService(repo repository.ServerRepository) ServerService {
	return &serverService{repo}
}

func (s *serverService) GetPaginatedServers(ctx context.Context, page, limit int, filters map[string][]string) (*models.PaginatedServers, error) {
	data, err := s.repo.GetServers(ctx, page, limit, filters)
	if err != nil {
		return nil, err
	}

	totalItems, err := s.repo.CountServers(ctx, filters)
	if err != nil {
		return nil, err
	}

	totalPages := (totalItems + limit - 1) / limit

	return &models.PaginatedServers{
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *serverService) SearchColumnValues(ctx context.Context, column, query string, limit int) ([]string, error) {
	return s.repo.SearchColumnValues(ctx, column, query, limit)
}

func (s *serverService) GetServerByID(ctx context.Context, id int) (*models.Server, error) {
	return s.repo.GetServerByID(ctx, id)
}