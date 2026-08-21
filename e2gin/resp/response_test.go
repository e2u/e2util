package resp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCtx(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func TestSuccessWithJSON(t *testing.T) {
	c, w := testCtx(t)
	SuccessWithJSON(c, Success, gin.H{"id": 1})
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.EqualValues(t, Success, body["code"])
	assert.Equal(t, "success", body["message"])
	assert.EqualValues(t, 1, body["id"])

	c, w = testCtx(t)
	SuccessWithJSON(c, Success, "hello")
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "hello", body["data"])

	c, w = testCtx(t)
	SuccessWithJSON(c, Success, nil)
	body = map[string]any{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	_, hasData := body["data"]
	assert.False(t, hasData)
}

func TestAboutWithJSON(t *testing.T) {
	c, w := testCtx(t)
	AboutWithJSON(c, BadRequest, errors.New("bad"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "bad", body["detail"])
	assert.True(t, c.IsAborted())

	c, w = testCtx(t)
	AboutWithJSON(c, NotFound, "missing")
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "missing", body["detail"])

	c, w = testCtx(t)
	AboutWithJSON(c, 12345, map[string]int{"n": 1})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRADataJSON(t *testing.T) {
	c, w := testCtx(t)
	c.Request.Header.Set(ConsumerHeader, ConsumerTypeRADataJSON)
	SuccessWithJSON(c, Success, []string{"a", "b"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2", w.Header().Get("X-Total-Count"))
	var items []string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	assert.Equal(t, []string{"a", "b"}, items)
}

func TestGetSliceLen(t *testing.T) {
	n, ok := getSliceLen([]int{1, 2, 3})
	assert.True(t, ok)
	assert.Equal(t, 3, n)
	n, ok = getSliceLen("nope")
	assert.False(t, ok)
	assert.Equal(t, 0, n)
}
