package product

import (
	"context"
	"microserviceArchWithGo/domain"
)

type Repository interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
}
