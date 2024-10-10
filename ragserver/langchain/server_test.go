package langchain

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (rs *testRagSuite) TestAddDocuments() {
	ctx := context.Background()

	rs.Run("Adds documents", func() {
		err := rs.rag.AddDocuments(ctx, []string{"You are mastermind.", "You are genius."})

		assert.Nil(rs.T(), err)
	})
}

func (rs *testRagSuite) TestAskQuestion() {
	ctx := context.Background()

	rs.Run("Generate response for the question", func() {
		err := rs.rag.AddDocuments(ctx, []string{"You are mastermind.", "You are genius."})

		assert.Nil(rs.T(), err)

		resp, err := rs.rag.AskQuestion(ctx, "Who are you?")

		assert.Nil(rs.T(), err)
		assert.NotEmpty(rs.T(), resp)
	})
}

type testRagSuite struct {
	suite.Suite
	rag *ragServer
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

	ragServer, err := New(ctx, log, os.Getenv("GEMINI_FLASH_API_KEY"))

	assert.Nil(rs.T(), err)
	rs.rag = ragServer
}
