package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"microserviceArchWithGo/pkg/config"
	_ "microserviceArchWithGo/pkg/log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("app starting...")

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Concurrency:  256 * 1024,
	})

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			zap.L().Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()
	zap.L().Info("Server started on port ", zap.String("port", appConfig.Port))

	gracefulShutdown(app)
}

func gracefulShutdown(app *fiber.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-sigChan
	zap.L().Info("Shutting down server...")

	if err := app.ShutdownWithTimeout(5 + time.Second); err != nil {
		zap.L().Error("Error during server shutdown", zap.Error(err))
	}
	zap.L().Info("Server gracefully stopped")
}

func https() {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.google.com", nil)
	if err != nil {
		zap.L().Error("Failed to create request", zap.Error(err))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		zap.L().Error("Failed to get google", zap.Error(err))
	}
	zap.L().Info("google response", zap.Int("status", resp.StatusCode))

}
