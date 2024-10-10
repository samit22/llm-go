package langchan

import (
	"cmp"
	"context"
	"fmt"
	"llm-rag/ragserver"
	"log/slog"
	"os"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

const (
	llmModel           = "gemini-1.5-flash"
	embeddingModelName = "text-embedding-004"
	collectionClass    = "Document"
)

type ragServer struct {
	log             *slog.Logger
	vectorDBClient  weaviate.Store
	generativeModel *googleai.GoogleAI
}

func New(ctx context.Context, log *slog.Logger, apiKey string) (*ragServer, error) {
	genClient, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultEmbeddingModel(embeddingModelName))
	if err != nil {
		return nil, fmt.Errorf("create google ai client %w", err)
	}

	emb, err := embeddings.NewEmbedder(genClient)
	if err != nil {
		return nil, fmt.Errorf("create embeddings %w", err)
	}

	weavClient, err := weaviate.New(
		weaviate.WithEmbedder(emb),
		weaviate.WithScheme("http"),
		weaviate.WithHost(cmp.Or(os.Getenv("WEAVIATE_DB_HOST"), "localhost:5555")),
		weaviate.WithIndexName(collectionClass),
	)
	if err != nil {
		return nil, fmt.Errorf("create weaviate client: %w", err)
	}

	return &ragServer{
		log:             log,
		vectorDBClient:  weavClient,
		generativeModel: genClient,
	}, nil
}

func (r *ragServer) Close() {
	// Close the clients if required
}

func (r *ragServer) AddDocuments(ctx context.Context, documents []string) error {
	var docs []schema.Document
	for _, doc := range documents {
		docs = append(docs, schema.Document{PageContent: doc})
	}
	_, err := r.vectorDBClient.AddDocuments(ctx, docs)
	if err != nil {
		return fmt.Errorf("adding documents to Weaviate: %w", err)
	}
	return nil
}

func (r *ragServer) AskQuestion(ctx context.Context, question string) (string, error) {
	results, err := r.vectorDBClient.SimilaritySearch(ctx, question, 4)
	if err != nil {
		return "", fmt.Errorf("querying weaviate similarity search: %w", err)
	}
	var resContents = make([]string, len(results))
	for i, res := range results {
		resContents[i] = res.PageContent
	}

	ragQuery := fmt.Sprintf(ragserver.Template, question, strings.Join(resContents, "\n"))
	llmModel := llms.WithModel(llmModel)
	res, err := llms.GenerateFromSinglePrompt(ctx, r.generativeModel, ragQuery, llmModel)
	if err != nil {
		return "", fmt.Errorf("gen model response: %w", err)
	}
	return res, nil
}
