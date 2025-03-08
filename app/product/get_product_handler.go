package product

import (
	"context"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"io"
	"microserviceArchWithGo/domain"
	"net/http"
	"time"
)

type GetProductRequest struct {
	ID string `json:"id" param:"id"`
}

type GetProductResponse struct {
	Product *domain.Product `json:"product"`
}

type GetProductHandler struct {
	repository Repository
	httpClient *retryablehttp.Client
	breaker    *gobreaker.CircuitBreaker
	httpServer string
}

func NewGetProductHandler(repository Repository, httpClient *retryablehttp.Client, httpServer string) *GetProductHandler {
	breakerSettings := gobreaker.Settings{
		Name:        "http-client",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRation := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRation >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			zap.L().Info("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}

	return &GetProductHandler{
		repository: repository,
		httpClient: httpClient,
		breaker:    gobreaker.NewCircuitBreaker(breakerSettings),
		httpServer: httpServer,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, h.httpServer+"/random-error", nil)

	if err != nil {
		return nil, err
	}

	retryableReq, err := retryablehttp.FromRequest(httpReq)
	if err != nil {
		return nil, err
	}

	resp, err := h.breaker.Execute(func() (interface{}, error) {
		return h.httpClient.Do(retryableReq)
	})

	if err != nil {
		return nil, err
	}

	httpResp := resp.(*http.Response)
	defer httpResp.Body.Close()
	if _, err = io.ReadAll(httpResp.Body); err != nil {
		return nil, err
	}

	product, err := h.repository.GetProduct(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &GetProductResponse{Product: product}, nil
}
