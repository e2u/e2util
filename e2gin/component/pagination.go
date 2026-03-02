package component

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strings"

	"github.com/e2u/e2util/e2strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Constants for query parameters and defaults
// 查詢參數和默認值的常量
const (
	NamePageSize       = "_ps"
	NamePageNumber     = "_pn"
	NameOrderField     = "_of"
	NameOrderDirection = "_od"
	ValueOrderByAsc    = "ASC"
	ValueOrderByDesc   = "DESC"
	DefaultOrderField  = "created_at"
	JSONItemsKey       = "items"
)

// PaginationResult holds paginated data
// PaginationResult 保存分頁數據
type PaginationResult[T any] struct {
	Items          T              `json:"items,omitempty"`
	Total          int64          `json:"total"`
	Pages          int            `json:"pages"`
	PrePage        int            `json:"pre_page"`
	Offset         int            `json:"offset"`
	Page           int            `json:"page"`
	CurrentPages   int            `json:"current_pages"`
	Html           *template.HTML `json:"html,omitempty"`
	OrderField     string         `json:"order_field,omitempty"`
	OrderDirection string         `json:"order_direction,omitempty"`
}

// PaginationOption defines options for pagination
// PaginationOption 定義分頁選項
type PaginationOption struct {
	DisablePagination   bool
	PrePage             int
	Page                int
	Offset              int
	OrderField          string
	OrderDirection      string
	DisableHtmlBar      bool
	DisableHeadCount    bool
	HeadCountName       string
	DisableContentRange bool
}

// BuildQueryUri constructs a URL with query parameters
// BuildQueryUri 構造帶有查詢參數的 URL
func BuildQueryUri(c *gin.Context, nv map[string]any, replaceQuery bool) string {
	qv := c.Request.URL.Query()
	if replaceQuery {
		qv = url.Values{}
	}
	for k, v := range nv {
		qv.Set(k, fmt.Sprintf("%v", v))
	}
	// Sanitize the URL to prevent injection
	// 消毒 URL 以防止注入
	return fmt.Sprintf("%s?%s", c.Request.URL.Path, url.QueryEscape(qv.Encode()))
}

// parsePaginationOptions parses and validates query parameters
// parsePaginationOptions 解析並驗證查詢參數
func parsePaginationOptions(c *gin.Context, opts ...*PaginationOption) (*PaginationOption, error) {
	opt := &PaginationOption{
		PrePage:        10,
		Page:           1,
		OrderField:     DefaultOrderField,
		OrderDirection: ValueOrderByDesc,
	}
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	}

	// Parse query parameters with validation
	// 解析查詢參數並進行驗證
	if v, ok := c.GetQuery(NamePageSize); ok {
		if pageSize, err := e2strconv.ParseInt[int](v); err == nil && pageSize > 0 {
			opt.PrePage = pageSize
		} else {
			return nil, fmt.Errorf("invalid page size: %s", v)
		}
	}
	if v, ok := c.GetQuery(NamePageNumber); ok {
		if pageNum, err := e2strconv.ParseInt[int](v); err == nil && pageNum > 0 {
			opt.Page = pageNum
		} else {
			return nil, fmt.Errorf("invalid page number: %s", v)
		}
	}
	if v, ok := c.GetQuery(NameOrderField); ok {
		if v != "" {
			opt.OrderField = v
		}
	}
	if v, ok := c.GetQuery(NameOrderDirection); ok {
		v = strings.ToUpper(v)
		if v == ValueOrderByAsc || v == ValueOrderByDesc {
			opt.OrderDirection = v
		} else {
			return nil, fmt.Errorf("invalid order direction: %s", v)
		}
	}

	// Ensure defaults for empty or invalid fields
	// 確保空或無效欄位的默認值
	if opt.OrderField == "" {
		opt.OrderField = DefaultOrderField
	}
	if opt.OrderDirection == "" {
		opt.OrderDirection = ValueOrderByDesc
	}
	if opt.PrePage <= 0 {
		opt.PrePage = 10
	}
	if opt.Page <= 0 {
		opt.Page = 1
	}

	return opt, nil
}

// calculatePagination computes pagination metadata
// calculatePagination 計算分頁元數據
func calculatePagination(opt *PaginationOption, totalCount int64) (offset int, pages int) {
	if opt.Page <= 1 {
		opt.Page = 1
		offset = 0
	} else {
		offset = (opt.Page - 1) * opt.PrePage
	}
	if opt.Offset > 0 {
		offset = opt.Offset
	}
	pages = int(math.Ceil(float64(totalCount) / float64(opt.PrePage)))
	return offset, pages
}

// generatePaginationHtml creates the HTML pagination bar
// generatePaginationHtml 創建 HTML 分頁欄
func generatePaginationHtml(c *gin.Context, prs *PaginationResult[[]any]) *template.HTML {
	if prs.Total <= 0 || prs.PrePage <= 0 {
		return nil
	}

	var liStr strings.Builder
	start := 1
	if prs.Page > 2 {
		start = prs.Page - 1
	}
	end := prs.Page + 1
	if prs.Page <= 3 {
		end = prs.Page + 2
	}
	if prs.Page == 1 && prs.Pages > 1 {
		end++
	}
	if prs.Page >= prs.Pages && prs.Page > 3 {
		start = prs.Page - 3
	}
	if start < 1 {
		start = 1
	}
	if end > prs.Pages {
		end = prs.Pages
	}

	for i := start; i <= end; i++ {
		if i == prs.Page {
			fmt.Fprintf(&liStr, `<li><span class="active">%d</span></li>`, i)
		} else {
			fmt.Fprintf(&liStr, `<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: i}, false), i)
		}
		if i < end {
			liStr.WriteString("\n")
		}
	}

	var templateBuilder strings.Builder
	templateBuilder.WriteString(`<div class="pagination-wrapper"><ul class="pagination">`)
	if prs.Page > 3 {
		fmt.Fprintf(&templateBuilder, `<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: 1}, false), 1)
	}
	templateBuilder.WriteString(liStr.String())
	if prs.Page < prs.Pages {
		fmt.Fprintf(&templateBuilder, `<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: prs.Pages}, false), prs.Pages)
	}
	templateBuilder.WriteString(`</ul></div>`)

	html := template.HTML(templateBuilder.String()) // #nosec G203
	return &html
}

// PaginationList paginates a list of items of type T
// PaginationList 分頁處理 T 類型的項目列表
func PaginationList[T any](c *gin.Context, model T, dbQuery *gorm.DB, opts ...*PaginationOption) (*PaginationResult[[]T], error) {
	if c == nil {
		return nil, fmt.Errorf("invalid context type")
	}

	// Parse and validate options
	// 解析並驗證選項
	opt, err := parsePaginationOptions(c, opts...)
	if err != nil {
		return nil, err
	}

	// Set up the query
	// 設置查詢
	dbQuery = dbQuery.Model(model)

	// Get total count
	// 獲取總數
	var totalCount int64
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	// 計算分頁元數據
	offset, pages := calculatePagination(opt, totalCount)

	// Build order string
	// 構建排序字符串
	var orderStr strings.Builder
	if strings.Contains(opt.OrderField, ",") {
		ofs := strings.Split(opt.OrderField, ",")
		for i, v := range ofs {
			fmt.Fprintf(&orderStr, "%s %s", v, opt.OrderDirection)
			if i < len(ofs)-1 {
				orderStr.WriteString(",")
			}
		}
	} else {
		fmt.Fprintf(&orderStr, "%s %s", opt.OrderField, opt.OrderDirection)
	}

	// Execute query
	// 執行查詢
	var ls []T
	if opt.DisablePagination {
		if err := dbQuery.Order(orderStr.String()).Find(&ls).Error; err != nil {
			return nil, err
		}

	} else {
		if err := dbQuery.Order(orderStr.String()).Limit(opt.PrePage).Offset(offset).Find(&ls).Error; err != nil {
			return nil, err
		}
	}

	if !opt.DisableHeadCount {
		if opt.HeadCountName == "" {
			opt.HeadCountName = "X-Total-Count"
		}
		c.Header("X-Total-Count", fmt.Sprintf("%d", totalCount))
	}

	if !opt.DisableContentRange {
		c.Header("Content-Range", fmt.Sprintf("items %d-%d/%d", offset, len(ls)-1, totalCount))
	}

	// Build result
	// 構建結果
	prs := &PaginationResult[[]T]{
		PrePage:        opt.PrePage,
		Items:          ls,
		Total:          totalCount,
		Pages:          pages,
		Page:           opt.Page,
		Offset:         offset,
		CurrentPages:   len(ls),
		OrderField:     opt.OrderField,
		OrderDirection: opt.OrderDirection,
	}

	// Generate HTML if not disabled
	// 如果未禁用，則生成 HTML
	if !opt.DisableHtmlBar {
		prsInterface := &PaginationResult[[]any]{
			PrePage:        prs.PrePage,
			Total:          prs.Total,
			Pages:          prs.Pages,
			Page:           prs.Page,
			Offset:         prs.Offset,
			CurrentPages:   prs.CurrentPages,
			OrderField:     prs.OrderField,
			OrderDirection: prs.OrderDirection,
		}
		prs.Html = generatePaginationHtml(c, prsInterface)
	}

	return prs, nil
}
