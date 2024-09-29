package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *handler) addDocumentsHandler(ctx *gin.Context) {
	var request AddDocumentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.log.Error(createMessage("Failed to parse request JSON %v", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.ragServer.AddDocuments(ctx, request.Documents)
	if err != nil {
		h.log.Error(createMessage("Failed to add documents %v", err))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"message": "Documents added successfully."})

}

func (h *handler) askQuestion(ctx *gin.Context) {
	var request AskQuestionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.log.Error(createMessage("Failed to parse request JSON %v", err))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	answer, err := h.ragServer.AskQuestion(ctx, request.Question)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"answer": answer})
}

func (h *handler) recover(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error(createMessage("Panic recovered %v", r))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()
	ctx.Next()
}

func (h *handler) accessLog(ctx *gin.Context) {
	log := h.log.With("log_type", "access")
	startTime := time.Now().UTC()
	ctx.Next()
	args := []interface{}{
		"request_on", startTime.Format("2006-01-02T15:04:05Z07:00"),
		"log_type", "access",
		"method", ctx.Request.Method,
		"params", ctx.Request.URL.Query().Encode(),
		"path", ctx.Request.URL.Path,
		"remote_addr", ctx.ClientIP(),
		"status", ctx.Writer.Status(),
		"content_length", ctx.Writer.Size(),
		"duration", time.Since(startTime).Milliseconds(),
	}
	if ctx.Errors != nil || ctx.Writer.Status() >= 500 {
		log.Error(ctx.Errors.String(), args...)
		return
	}
	log.Info("", args...)
}
