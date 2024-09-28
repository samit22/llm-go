package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	t.Log("When api key is not set it exits with error code")
	{
		if os.Getenv("GEMINI_NOT_SET") == "1" {
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
