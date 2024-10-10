package ragserver

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"google.golang.org/api/option"
)

const (
	llmModel           = "gemini-1.5-flash"
	embeddingModelName = "text-embedding-004"
	collectionClass    = "Document"
)

type server struct {
	log             *slog.Logger
	genClient       *genai.Client
	vectorDBClient  *weaviate.Client
	generativeModel *genai.GenerativeModel
	embedModel      *genai.EmbeddingModel
}

func New(ctx context.Context, log *slog.Logger, apiKey string) (*server, error) {
	genClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create genai client %w", err)
	}
	weavClient, err := weaviate.NewClient(weaviate.Config{
		Host:   cmp.Or(os.Getenv("WEAVIATE_DB_HOST"), "localhost:5555"),
		Scheme: "http",
	})
	if err != nil {
		return nil, fmt.Errorf("create weaviate client: %w", err)
	}
	err = checkWeaviateCollection(ctx, weavClient, collectionClass)
	if err != nil {
		return nil, fmt.Errorf("check weaviate collection: %w", err)
	}

	return &server{
		log:             log,
		vectorDBClient:  weavClient,
		genClient:       genClient,
		generativeModel: genClient.GenerativeModel(llmModel),
		embedModel:      genClient.EmbeddingModel(embeddingModelName),
	}, nil
}

// Close disconnects clients that requires closing
func (rag *server) Close() {
	rag.genClient.Close()
}

// AddDocuments adds new document to the knowledge base
func (rag *server) AddDocuments(ctx context.Context, documents []string) error {
	batch := rag.embedModel.NewBatch()
	for _, doc := range documents {
		batch.AddContent(genai.Text(doc))
	}
	embedModelResp, err := rag.embedModel.BatchEmbedContents(ctx, batch)
	if err != nil {
		return fmt.Errorf("embedding documents: %w", err)
	}
	rag.log.Info("Embeddings generated successfully")
	if len(embedModelResp.Embeddings) != len(documents) {
		return fmt.Errorf("expected %d embeddings, got %d", len(documents), len(embedModelResp.Embeddings))
	}

	vectorObjs := make([]*models.Object, len(documents))
	for i, doc := range documents {
		vectorObjs[i] = &models.Object{
			Class: collectionClass,
			Properties: map[string]any{
				"text": doc,
			},
			Vector: embedModelResp.Embeddings[i].Values,
		}
	}

	_, err = rag.vectorDBClient.Batch().ObjectsBatcher().WithObjects(vectorObjs...).Do(ctx)
	if err != nil {
		return fmt.Errorf("adding documents to Weaviate: %w", err)
	}
	rag.log.Info("Documents added to weaviate successfully")
	return nil

}

// Ask question to send the query
func (rag *server) AskQuestion(ctx context.Context, question string) (string, error) {
	embedModelResp, err := rag.embedModel.EmbedContent(ctx, genai.Text(question))
	if err != nil {
		return "", fmt.Errorf("embedding question: %w", err)
	}

	grahpQ := rag.vectorDBClient.GraphQL()
	result, err := grahpQ.Get().
		WithNearVector(grahpQ.NearVectorArgBuilder().WithVector(embedModelResp.Embedding.Values)).
		WithClassName(collectionClass).
		WithFields(graphql.Field{Name: "text"}).
		WithLimit(4).
		Do(ctx)
	if err != nil {
		return "", fmt.Errorf("querying weaviate: %w", err)
	}
	if len(result.Errors) > 0 {
		var err strings.Builder
		for _, e := range result.Errors {
			err.WriteString(e.Message)
			err.WriteString("\t")
		}
		return "", fmt.Errorf("weaviate query error: %s", err.String())
	}

	vectorContexts, err := extractGraphResult(result)
	if err != nil {
		return "", fmt.Errorf("decoding weaviate results: %w", err)
	}

	ragQuery := fmt.Sprintf(Template, question, strings.Join(vectorContexts, "\n"))
	llmResp, err := rag.generativeModel.GenerateContent(ctx, genai.Text(ragQuery))
	if err != nil {
		return "", fmt.Errorf("generating response from LLM: %w", err)
	}

	if len(llmResp.Candidates) < 1 {
		return "", fmt.Errorf("unexpected candidates count %d", len(llmResp.Candidates))
	}

	var respContents []string
	for _, part := range llmResp.Candidates[0].Content.Parts {
		if pt, ok := part.(genai.Text); ok {
			respContents = append(respContents, string(pt))
		} else {
			log.Printf("bad type of part: %v", pt)
			return "", fmt.Errorf("unexpected content part type %T", pt)
		}
	}
	return strings.Join(respContents, "\n"), nil
}

func extractGraphResult(result *models.GraphQLResponse) ([]string, error) {
	var graphResp GraphQLResponse
	byteData, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}
	err = json.Unmarshal(byteData, &graphResp)
	if err != nil {
		return nil, fmt.Errorf("create data: %w", err)
	}
	var out []string
	for _, doc := range graphResp.Get.Document {
		out = append(out, doc.Text)
	}
	return out, nil
}

func checkWeaviateCollection(ctx context.Context, client *weaviate.Client, collClass string) error {
	coll := &models.Class{
		Class:      collClass,
		Vectorizer: "none",
	}
	exists, err := client.Schema().ClassExistenceChecker().WithClassName(coll.Class).Do(ctx)
	if err != nil {
		return fmt.Errorf("weaviate class check error: %w", err)
	}
	if exists {
		return nil
	}
	err = client.Schema().ClassCreator().WithClass(coll).Do(ctx)
	if err != nil {
		return fmt.Errorf("weaviate create class error: %w", err)
	}

	return nil
}
