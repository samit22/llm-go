package ragserver

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Log("Initialize server")
	ctx := context.Background()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	srv, err := New(ctx, log, "testKey")

	assert.Nil(t, err)
	assert.NotNil(t, srv)
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
		assert.Contains(rs.T(), strings.ToLower(resp), "mastermind")
	})
}
