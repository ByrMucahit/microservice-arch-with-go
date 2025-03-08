package couchbase

import (
	"context"
	"errors"
	gocbopentelemetry "github.com/couchbase/gocb-opentelemetry"
	"github.com/couchbase/gocb/v2"
	sdktrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"microserviceArchWithGo/domain"
	"time"
)

type CouchBaseRepository struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	tp      *sdktrace.TracerProvider
	tracer  *gocbopentelemetry.OpenTelemetryRequestTracer
}

func NewCouchbaseRepository(tp *sdktrace.TracerProvider, couchbaseUrl string, username string, password string) *CouchBaseRepository {
	tracer := gocbopentelemetry.NewOpenTelemetryRequestTracer(tp)
	cluster, err := gocb.Connect(couchbaseUrl, gocb.ClusterOptions{
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: 3 * time.Second,
			KVTimeout:      3 * time.Second,
			QueryTimeout:   3 * time.Second,
		},
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
		Transcoder: gocb.NewJSONTranscoder(),
		Tracer:     tracer,
	})
	if err != nil {
		zap.L().Fatal("Failed to connect to couchbase", zap.Error(err))
	}

	bucket := cluster.Bucket("products")
	bucket.WaitUntilReady(3*time.Second, &gocb.WaitUntilReadyOptions{})

	return &CouchBaseRepository{
		cluster: cluster,
		bucket:  bucket,
		tracer:  tracer,
	}
}

func (r *CouchBaseRepository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	ctx, span := r.tracer.Wrapped().Start(ctx, "GetProduct")
	defer span.End()

	data, err := r.bucket.DefaultCollection().Get(id, &gocb.GetOptions{
		Timeout:    3 * time.Second,
		Context:    ctx,
		ParentSpan: gocbopentelemetry.NewOpenTelemetryRequestSpan(ctx, span),
	})

	if err != nil {
		if errors.Is(err, gocb.ErrDocumentNotFound) {
			return nil, errors.New("product not found")
		}

		zap.L().Error("Failed to get product", zap.Error(err))
		return nil, err
	}

	var product domain.Product

}
