package ragserver

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"google.golang.org/api/option"
)

type testRagSuite struct {
	suite.Suite
	rag *server
}

func TestRagServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(testRagSuite))
}

func (rs *testRagSuite) SetupSuite() {
	ctx := context.Background()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	genClient, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_FLASH_API_KEY")))
	assert.Nil(rs.T(), err)

	vectorDBClient, err := weaviate.NewClient(weaviate.Config{
		Host:   cmp.Or(os.Getenv("WEAVIATE_DB_HOST"), "localhost:5555"),
		Scheme: "http",
	})
	assert.Nil(rs.T(), err)
	rs.rag = &server{
		log:             log,
		genClient:       genClient,
		vectorDBClient:  vectorDBClient,
		generativeModel: genClient.GenerativeModel("gemini-1.5-flash"),
		embedModel:      genClient.EmbeddingModel("text-embedding-004"),
	}
	err = checkWeaviateCollection(ctx, vectorDBClient, collectionClass)
	assert.Nil(rs.T(), err)
}

func (rs *testRagSuite) TearDownSuite() {
	rs.rag.genClient.Close()
}
