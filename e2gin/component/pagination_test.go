package component

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testGinContext(t *testing.T, rawURL string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	require.NoError(t, err)
	c.Request = req
	return c
}

func TestBuildQueryUri(t *testing.T) {
	c := testGinContext(t, "/list?foo=bar")
	got := BuildQueryUri(c, map[string]any{NamePageNumber: 2}, false)
	assert.Equal(t, "/list?_pn=2&foo=bar", got)

	got = BuildQueryUri(c, map[string]any{"q": "a b"}, true)
	assert.Equal(t, "/list?q=a+b", got)

	c = testGinContext(t, "/list")
	got = BuildQueryUri(c, map[string]any{}, false)
	assert.Equal(t, "/list", got)
}

func TestValidOrderFields(t *testing.T) {
	assert.True(t, validOrderFields("created_at"))
	assert.True(t, validOrderFields("users.created_at"))
	assert.True(t, validOrderFields("id,created_at"))
	assert.True(t, validOrderFields(" users.id , created_at "))
	assert.False(t, validOrderFields("created_at;drop table x"))
	assert.False(t, validOrderFields("id desc"))
	assert.False(t, validOrderFields(""))
	assert.False(t, validOrderFields("1id"))
}

func TestParsePaginationOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		c := testGinContext(t, "/items")
		opt, err := parsePaginationOptions(c)
		require.NoError(t, err)
		assert.Equal(t, 10, opt.PrePage)
		assert.Equal(t, 1, opt.Page)
		assert.Equal(t, DefaultOrderField, opt.OrderField)
		assert.Equal(t, ValueOrderByDesc, opt.OrderDirection)
	})

	t.Run("query values", func(t *testing.T) {
		c := testGinContext(t, "/items?_ps=20&_pn=3&_of=name&_od=ASC")
		opt, err := parsePaginationOptions(c)
		require.NoError(t, err)
		assert.Equal(t, 20, opt.PrePage)
		assert.Equal(t, 3, opt.Page)
		assert.Equal(t, "name", opt.OrderField)
		assert.Equal(t, ValueOrderByAsc, opt.OrderDirection)
	})

	t.Run("rejects sql injection order field", func(t *testing.T) {
		c := testGinContext(t, "/items?_of=created_at%20ASC")
		_, err := parsePaginationOptions(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid order field")
	})

	t.Run("rejects option order field", func(t *testing.T) {
		c := testGinContext(t, "/items")
		_, err := parsePaginationOptions(c, &PaginationOption{OrderField: "id desc"})
		require.Error(t, err)
	})

	t.Run("invalid page size", func(t *testing.T) {
		c := testGinContext(t, "/items?_ps=nope")
		_, err := parsePaginationOptions(c)
		require.Error(t, err)
	})
}

func TestCalculatePagination(t *testing.T) {
	opt := &PaginationOption{Page: 3, PrePage: 10}
	offset, pages := calculatePagination(opt, 25)
	assert.Equal(t, 20, offset)
	assert.Equal(t, 3, pages)

	opt = &PaginationOption{Page: 1, PrePage: 10, Offset: 7}
	offset, pages = calculatePagination(opt, 25)
	assert.Equal(t, 7, offset)
	assert.Equal(t, 3, pages)
}

func TestGeneratePaginationHtml(t *testing.T) {
	c := testGinContext(t, "/list?foo=1")
	assert.Nil(t, generatePaginationHtml(c, &PaginationResult[[]any]{Total: 0, PrePage: 10}))

	html := generatePaginationHtml(c, &PaginationResult[[]any]{
		Total: 100, PrePage: 10, Page: 1, Pages: 10,
	})
	require.NotNil(t, html)
	assert.Contains(t, string(*html), `class="active"`)
	assert.Contains(t, string(*html), `_pn=2`)

	html = generatePaginationHtml(c, &PaginationResult[[]any]{
		Total: 100, PrePage: 10, Page: 5, Pages: 10,
	})
	require.NotNil(t, html)
	assert.Contains(t, string(*html), `_pn=1`)
	assert.Contains(t, string(*html), `_pn=10`)
}

type pageRow struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time
}

func testDB(t *testing.T, n int) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&pageRow{}))
	for i := 0; i < n; i++ {
		require.NoError(t, db.Create(&pageRow{Name: fmt.Sprintf("n%02d", i)}).Error)
	}
	return db
}

func TestPaginationList(t *testing.T) {
	db := testDB(t, 25)

	t.Run("nil context", func(t *testing.T) {
		_, err := PaginationList(nil, pageRow{}, db)
		require.Error(t, err)
	})

	t.Run("pages results and headers", func(t *testing.T) {
		c := testGinContext(t, "/items?_ps=10&_pn=2&_of=id&_od=ASC")
		prs, err := PaginationList(c, pageRow{}, db)
		require.NoError(t, err)
		assert.Equal(t, int64(25), prs.Total)
		assert.Equal(t, 3, prs.Pages)
		assert.Equal(t, 10, prs.CurrentPages)
		assert.Equal(t, 10, prs.Offset)
		assert.Equal(t, "25", c.Writer.Header().Get("X-Total-Count"))
		assert.Contains(t, c.Writer.Header().Get("Content-Range"), "items 10-")
		require.NotNil(t, prs.Html)
	})

	t.Run("disable pagination returns all", func(t *testing.T) {
		c := testGinContext(t, "/items?_of=id&_od=ASC")
		prs, err := PaginationList(c, pageRow{}, db, &PaginationOption{
			DisablePagination:   true,
			DisableHtmlBar:      true,
			DisableHeadCount:    true,
			DisableContentRange: true,
			OrderField:          "id",
			OrderDirection:      ValueOrderByAsc,
		})
		require.NoError(t, err)
		assert.Len(t, prs.Items, 25)
		assert.Nil(t, prs.Html)
	})

	t.Run("comma order fields", func(t *testing.T) {
		c := testGinContext(t, "/items?_of=id,name&_od=ASC")
		prs, err := PaginationList(c, pageRow{}, db, &PaginationOption{DisableHtmlBar: true})
		require.NoError(t, err)
		assert.Equal(t, "id,name", prs.OrderField)
		assert.NotEmpty(t, prs.Items)
	})

	t.Run("invalid query", func(t *testing.T) {
		c := testGinContext(t, "/items?_of=id%3Bdrop")
		_, err := PaginationList(c, pageRow{}, db)
		require.Error(t, err)
	})
}
