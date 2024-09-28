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
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	log := slog.New(jsonHandler)

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
	handler := handler{
		log:       log,
		ragServer: rag,
	}

	engine := gin.New()
	engine.Use(handler.recover)

	engine.POST("/add-documents", handler.addDocumentsHandler)
	engine.POST("/ask", handler.askQuestion)
	engine.Run(":" + cmp.Or(os.Getenv("RAG_PORT"), "5000"))

}

func createMessage(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}
