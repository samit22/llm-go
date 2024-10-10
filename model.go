package main

import (
	"context"
	"log/slog"
)

const (
	rawSDK       = "raw"
	langchainSDK = "langchan"
)

type RagServer interface {
	AddDocuments(ctx context.Context, documents []string) error
	AskQuestion(ctx context.Context, question string) (string, error)
	Close()
}

type handler struct {
	log       *slog.Logger
	ragServer RagServer
}

type AddDocumentRequest struct {
	Documents []string `json:"documents" binding:"required,dive,required"`
}

type AskQuestionRequest struct {
	Question string `json:"question" binding:"required"`
}
