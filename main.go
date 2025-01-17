package main

import (
	"ac/bootstrap"
	"ac/bootstrap/logger"
	"ac/controller/auth"
	"ac/controller/permission"
	"ac/controller/resource"
	"ac/controller/role"

	"ac/controller/system"
	"ac/controller/user"

	"ac/controller/user_role"
	"ac/custom/output"
	"ac/custom/validator"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize the system
	err := bootstrap.Initialize()
	if err != nil {
		panic(fmt.Errorf("failed to initialize, err: %w", err))
	}

	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.NewCustomValidator()
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			kv := map[string]interface{}{
				"latency":   v.Latency,
				"remote_ip": v.RemoteIP,
				"status":    v.Status,
			}
			if v.Error == nil {
				logger.LogWith(c, logger.LevelInfo, "success", kv)
			} else {
				kv["error"] = v.Error.Error()
				logger.LogWith(c, logger.LevelError, "failure", kv)
			}
			return v.Error
		},
	}))
	e.Use(middleware.Recover())

	system.RegisterRoutes(e.Group("/system"))
	user.RegisterRoutes(e.Group("/user"))
	role.RegisterRoutes(e.Group("/role"))
	resource.RegisterRoutes(e.Group("/resource"))
	user_role.RegisterRoutes(e.Group("/user-role"))
	permission.RegisterRoutes(e.Group("/permission"))
	auth.RegisterRoutes(e.Group("/auth"))

	// Output all routes
	printRoutes(e)

	e.GET("/", func(c echo.Context) error {
		return output.Success(c, nil)
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start server, err: %w", err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("failed to shutdown server, err: %w", err))
	}
}

func printRoutes(e *echo.Echo) {
	routes := e.Routes()
	fmt.Println("Registered Routes:")
	for _, route := range routes {
		fmt.Printf("Method: %-6s | Path: %-30s\n", route.Method, route.Path)
	}
}
