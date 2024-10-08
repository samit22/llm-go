package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := slog.Default()
	t.Log("Add Documents Handler")
	{
		t.Run("Invalid request body returns 400", func(t *testing.T) {
			h := handler{
				log: log,
			}
			ginCtx, mockResp := mockGinContext()

			h.addDocumentsHandler(ginCtx)

			assert.Equal(t, http.StatusBadRequest, mockResp.Code)
		})

		t.Run("Error on adding documents returns 500", func(t *testing.T) {
			mockRag := &mockRagServer{}
			h := handler{
				log:       log,
				ragServer: mockRag,
			}
			ginCtx, mockResp := mockGinContext()
			ginCtx.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(`{"documents": ["You are mastermind.", "You are genius." ]}`)))
			mockRag.On("AddDocuments", mock.Anything, mock.Anything).Return(assert.AnError)

			h.addDocumentsHandler(ginCtx)

			assert.Equal(t, http.StatusInternalServerError, mockResp.Code)
		})

		t.Run("Responds with success message on successful addition", func(t *testing.T) {
			mockRag := &mockRagServer{}
			h := handler{
				log:       log,
				ragServer: mockRag,
			}
			ginCtx, mockResp := mockGinContext()
			ginCtx.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(`{"documents": ["You are mastermind.", "You are genius." ]}`)))
			mockRag.On("AddDocuments", mock.Anything, mock.Anything).Return(nil)

			h.addDocumentsHandler(ginCtx)

			assert.Equal(t, http.StatusOK, mockResp.Code)
		})
	}
	t.Log("Ask Question Handler")
	{
		t.Run("Invalid request body returns 400", func(t *testing.T) {
			h := handler{
				log: log,
			}
			ginCtx, mockResp := mockGinContext()

			h.askQuestion(ginCtx)

			assert.Equal(t, http.StatusBadRequest, mockResp.Code)
		})

		t.Run("Error on asking question returns 500", func(t *testing.T) {
			mockRag := &mockRagServer{}
			h := handler{
				log:       log,
				ragServer: mockRag,
			}
			ginCtx, mockResp := mockGinContext()
			ginCtx.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(`{"question": "Who are you?"}`)))
			mockRag.On("AskQuestion", mock.Anything, mock.Anything).Return("", assert.AnError)

			h.askQuestion(ginCtx)

			assert.Equal(t, http.StatusInternalServerError, mockResp.Code)
		})

		t.Run("Responds with answer on successful question", func(t *testing.T) {
			mockRag := &mockRagServer{}
			h := handler{
				log:       log,
				ragServer: mockRag,
			}
			ginCtx, mockResp := mockGinContext()
			ginCtx.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(`{"question": "Who are you?"}`)))
			mockRag.On("AskQuestion", mock.Anything, mock.Anything).Return("I am mastermind.", nil)

			h.askQuestion(ginCtx)

			assert.Equal(t, http.StatusOK, mockResp.Code)
		})
	}
	t.Log("Recover middleware")
	{
		t.Run("Does not crash the program for panics", func(t *testing.T) {
			w := httptest.NewRecorder()
			ginCtx, engine := gin.CreateTestContext(w)

			h := handler{
				log: log,
			}

			ginCtx.Request = &http.Request{
				Method: http.MethodGet,
				Header: make(http.Header),
				URL: &url.URL{
					Path: "/panic",
				},
			}

			engine.Use(h.recover)
			engine.GET("/panic", func(ctx *gin.Context) {
				panic("Panic")
			})

			engine.ServeHTTP(w, ginCtx.Request)
		})
	}
	t.Log("Access log middleware")
	{
		var lasLog = func(rawData string) map[string]interface{} {
			var data map[string]interface{}
			strSplit := strings.Split(rawData, "/n")
			if len(strSplit) == 0 {
				return data
			}

			lastLog := strSplit[len(strSplit)-1]
			json.Unmarshal([]byte(lastLog), &data)
			return data
		}

		type testData struct {
			name       string
			respStatus int

			expectedLevel string
		}
		var tests = []testData{
			{
				name:          "500 status are recorded as error",
				respStatus:    500,
				expectedLevel: "ERROR",
			},
			{
				name:          "non 500 status are recorded as info",
				respStatus:    200,
				expectedLevel: "INFO",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ginCtx, engine := gin.CreateTestContext(w)
				var buf bytes.Buffer
				log := slog.New(slog.NewJSONHandler(&buf, nil))

				h := handler{
					log: log,
				}

				engine.Use(h.accessLog)
				engine.GET("/app", func(ctx *gin.Context) {
					ctx.JSON(tt.respStatus, gin.H{"message": "some message"})
				})

				ginCtx.Request = &http.Request{
					Method: http.MethodGet,
					Header: make(http.Header),
					URL: &url.URL{
						Path: "/app",
					},
				}
				engine.ServeHTTP(w, ginCtx.Request)

				data := lasLog(buf.String())
				assert.Equal(t, tt.expectedLevel, data["level"])

			})
		}
	}
}

type mockRagServer struct {
	mock.Mock
}

func (m *mockRagServer) AddDocuments(ctx context.Context, docs []string) error {
	args := m.Called(ctx, docs)

	return args.Error(0)
}

func (m *mockRagServer) AskQuestion(ctx context.Context, question string) (string, error) {
	args := m.Called(ctx, question)

	return args.String(0), args.Error(1)
}

func (m *mockRagServer) Close() {
	m.Called()
}

func mockGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}
	return ginCtx, w
}
