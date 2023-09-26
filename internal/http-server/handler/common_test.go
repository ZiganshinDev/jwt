package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderJSON(t *testing.T) {
	w := httptest.NewRecorder()

	testData := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, World!",
	}

	err := renderJSON(w, testData)
	require.NoError(t, err)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expectedContentType := "application/json"
	contentType := w.Header().Get("Content-Type")
	require.Equal(t, contentType, expectedContentType)

	var response struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	expectedMessage := "Hello, World!"
	require.Equal(t, response.Message, expectedMessage)
}

func TestRenderJSONError(t *testing.T) {
	w := httptest.NewRecorder()

	testData := struct {
		Message string `json:"message"`
	}{
		Message: "Err",
	}

	err := renderJSON(w, testData)
	_ = err

	expectedContentType := "Err"
	contentType := w.Header().Get("Content-Type")
	require.NotEqual(t, contentType, expectedContentType)

	var response struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	expectedMessage := "Hello, World!"
	require.NotEqual(t, response.Message, expectedMessage)
}

func TestGetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Add("MyHeader", "MyValue")

	headerName := "MyHeader"
	expectedValue := "MyValue"
	value, err := getHeader(req, headerName)
	require.NoError(t, err)
	require.Equal(t, value, expectedValue)
}

func TestGetHeaderError(t *testing.T) {
	req := httptest.NewRequest("GET", "/auth", nil)

	missingHeaderName := ""
	_, err := getHeader(req, missingHeaderName)
	require.Error(t, err)
}
