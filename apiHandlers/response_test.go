package apiHandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupResponseRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestInitialize(t *testing.T) {
	rf := Initialize()
	assert.NotNil(t, rf)
	assert.IsType(t, &ResponseFactory{}, rf)
}

func TestCreateErrorResponse(t *testing.T) {
	rf := Initialize()

	testCases := []struct {
		name           string
		errCode        string
		expectedStatus int
		expectedCode   string
		expectedError  string
	}{
		{
			name:           "Valid error code",
			errCode:        "GEN_00002",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "GEN_00002",
			expectedError:  "Invalid request body",
		},
		{
			name:           "Invalid error code",
			errCode:        "INVALID_ERROR_CODE",
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "GEN_99999",
			expectedError:  "Internal server error.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupResponseRouter()
			router.GET("/test", func(c *gin.Context) {
				rf.CreateErrorResponse(c, tc.errCode)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedCode, response.Code)
			assert.Equal(t, tc.expectedError, response.Error)
		})
	}
}

func TestCreateErrorResponsef(t *testing.T) {
	rf := Initialize()

	testCases := []struct {
		name           string
		errCode        string
		args           []interface{}
		expectedStatus int
		expectedCode   string
		expectedError  string
	}{
		{
			name:           "Valid error code with formatting",
			errCode:        "GEN_00003",
			args:           []interface{}{"test_field"},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "GEN_00003",
			expectedError:  "Missing field in body: test_field",
		},
		{
			name:           "Invalid error code",
			errCode:        "INVALID_ERROR_CODE",
			args:           []interface{}{},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "GEN_99999",
			expectedError:  "Internal server error.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupResponseRouter()
			router.GET("/test", func(c *gin.Context) {
				rf.CreateErrorResponsef(c, tc.errCode, tc.args...)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedCode, response.Code)
			assert.Equal(t, tc.expectedError, response.Error)
		})
	}
}

func TestCreateSuccessResponse(t *testing.T) {
	rf := Initialize()

	testCases := []struct {
		name           string
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid data",
			data:           map[string]string{"key": "value"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"key":"value"}`,
		},
		{
			name:           "Nil data",
			data:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{}`,
		},
		{
			name:           "Empty slice",
			data:           []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupResponseRouter()
			router.GET("/test", func(c *gin.Context) {
				rf.CreateOKResponse(c, tc.data)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.String())
		})
	}
}
