package ragserver

import (
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
		Host:   "localhost:5555",
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
}

func (rs *testRagSuite) TearDownSuite() {
	rs.rag.genClient.Close()
}

func (rs *testRagSuite) SetupSubTest() {
	ctx := context.Background()
	err := checkWeaviateCollection(ctx, rs.rag.vectorDBClient, collectionClass)
	rs.Assert().Nil(err)
}
