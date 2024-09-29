package main

import (
	"cmp"
	"context"
	"fmt"
	"llm-rag/ragserver"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()
	log := createLogger().With("log_type", "application")
	geminiKey := os.Getenv("GEMINI_FLASH_API_KEY")
	if geminiKey == "" {
		log.Error("GEMINI_FLASH_API_KEY is not set")
		os.Exit(1)
	}

	rag, err := ragserver.New(ctx, log, geminiKey)
	if err != nil {
		log.Error(createMessage("failed to initialize RAG server: %v", err))
		os.Exit(1)
	}
	defer rag.Close()
	handler := &handler{
		log:       log,
		ragServer: rag,
	}
	engine := setupHTTPServer(handler)
	engine.Run(":" + cmp.Or(os.Getenv("RAG_PORT"), "5000"))
}

func setupHTTPServer(h *handler) *gin.Engine {
	engine := gin.New()
	engine.Use(h.recover, h.accessLog)
	engine.POST("/add-documents", h.addDocumentsHandler)
	engine.POST("/ask", h.askQuestion)
	h.log.Info("Starting server on port 5000")
	return engine
}

func createMessage(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}

func createLogger() *slog.Logger {
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	log := slog.New(jsonHandler)
	return log
}
