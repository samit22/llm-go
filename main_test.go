package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	t.Log("When api key is not set it exits with error code")
	{
		if os.Getenv("GEMINI_NOT_SET") == "1" {
			t.Setenv("GEMINI_FLASH_API_KEY", "")
			main()
			return
		}
		opErr := bytes.NewBuffer(nil)
		cmd := exec.Command(os.Args[0], "-test.run=TestMain")
		cmd.Env = append(os.Environ(), "GEMINI_NOT_SET=1")
		cmd.Stderr = opErr
		err := cmd.Run()

		e, ok := err.(*exec.ExitError)
		assert.True(t, ok)
		assert.False(t, e.Success())

		assert.Contains(t, opErr.String(), "GEMINI_FLASH_API_KEY is not set")
	}
}

func TestSetupHTTPServer(t *testing.T) {
	t.Log("Test setupHTTPServer")
	{
		log := createLogger().With("log_type", "application")
		h := &handler{
			log: log,
		}

		engine := setupHTTPServer(h)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/add-documents", bytes.NewBuffer([]byte(`{"documents": ["You are mastermind.", "You are genius." ]}`)))
		engine.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/ask", bytes.NewBuffer([]byte(`{"question": "What is the meaning of life?"}`)))
		engine.ServeHTTP(w, req)
		assert.Equal(t, 500, w.Code)
	}
}
