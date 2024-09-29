package ragserver

import (
	"context"
	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
)

func (rs *testRagSuite) New() {
	rs.Run("Initializes clients", func() {
		ctx := context.Background()
		log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		srv, err := New(ctx, log, os.Getenv("GEMINI_FLASH_API_KEY"))

		assert.Nil(rs.T(), err)
		assert.NotNil(rs.T(), srv)
	})
}

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
