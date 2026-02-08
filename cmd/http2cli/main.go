package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"http2cli/internal/config"
	"http2cli/internal/executor"
	"http2cli/internal/handler"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Loaded %d tool(s) from configuration", len(cfg.Tools))

	// Initialize executor
	exec := executor.New(cfg.Security.AllowedCommands, cfg.Security.MaxConcurrentCommands)

	// Create router
	mux := http.NewServeMux()

	// Register standard endpoints
	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("GET /ready", handler.ReadyHandler(cfg, exec))
	mux.HandleFunc("GET /api/tools", handler.ToolsHandler(cfg))

	// Register tool endpoints
	for i := range cfg.Tools {
		tool := &cfg.Tools[i]
		h := handler.NewToolHandler(tool, exec)

		pattern := fmt.Sprintf("%s %s", tool.Method, tool.Endpoint)
		mux.Handle(pattern, h)
		log.Printf("Registered endpoint: %s", pattern)
	}

	// Apply middleware
	var h http.Handler = mux
	h = handler.APIKeyMiddleware(cfg.Security.APIKey, h)
	h = handler.LoggingMiddleware(h)

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server
	go func() {
		log.Printf("Starting http2cli on %s", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
