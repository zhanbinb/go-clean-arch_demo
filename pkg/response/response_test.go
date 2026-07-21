package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
)

func init() { gin.SetMode(gin.TestMode) }

func TestOK(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	OK(c, gin.H{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp Resp
	require := assert.New(t)
	require.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(0, resp.Code)
	require.Equal("OK", resp.Message)
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Created(c, gin.H{"id": 1})

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, errcode.ErrNotFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp Resp
	require := assert.New(t)
	require.NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(int(errcode.ErrNotFound.Code), resp.Code)
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
