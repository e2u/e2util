package e2webapp

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

const (
	NamePageSize       = "_ps"
	NamePageNumber     = "_pn"
	NameOrderField     = "_of"
	NameOrderDirection = "_od"

	ValueOrderByAsc  = "ASC"
	ValueOrderByDesc = "DESC"

	DefaultOrderField = "pkaid"
)

type PaginationResult struct {
	Items            any    `json:"items,omitempty"`
	TotalCount       int64  `json:"total_count"`
	TotalPages       int    `json:"total_page"`
	PageSize         int    `json:"page_size"`
	CurrentPageCount int    `json:"current_page_count"`
	PageNumber       int    `json:"page_number"`
	PaginationHtml   any    `json:"pagination_html,omitempty"`
	OrderField       string `json:"order_field,omitempty"`
	OrderDirection   string `json:"order_direction,omitempty"`
}

/**
example:

prs, err := e2webapp.PaginationList(c, &gallery.PhotoMate{}, cc.RW())
*/

func BuildQueryUri(c *gin.Context, nv map[string]any, replaceQuery bool) string {
	qv := c.Request.URL.Query()
	if replaceQuery {
		qv = url.Values{}
	}
	for k, v := range nv {
		qv.Set(k, fmt.Sprintf("%v", v))
	}
	return fmt.Sprintf("%s?%s", c.Request.URL.Path, qv.Encode())
}

type PaginationOption struct {
	PageSize       int
	PageNumber     int
	OrderField     string // _of=pkaid | _of=size | _of=created_at
	OrderDirection string // _or=ASC | _or=DESC
}

func PaginationList[T any](c *gin.Context, model T, dbQuery *gorm.DB, opts ...*PaginationOption) (*PaginationResult, error) {
	dbQuery = dbQuery.Model(model)
	var opt *PaginationOption

	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt == nil {
		opt = &PaginationOption{
			PageSize:       10,
			PageNumber:     1,
			OrderDirection: ValueOrderByDesc,
			OrderField:     DefaultOrderField,
		}
	} else {
		if opt.PageSize <= 0 {
			opt.PageSize = 10
		}
		if opt.PageNumber <= 0 {
			opt.PageNumber = 1
		}

		if opt.OrderField == "" {
			opt.OrderField = DefaultOrderField
		}

		if opt.OrderDirection == "" {
			opt.OrderDirection = ValueOrderByDesc
		}
	}

	if v, ok := c.GetQuery(NamePageSize); ok {
		opt.PageSize = e2strconv.MustParseInt(v)
	}

	if v, ok := c.GetQuery(NamePageNumber); ok {
		opt.PageNumber = e2strconv.MustParseInt(v)
	}

	if v, ok := c.GetQuery(NameOrderField); ok {
		opt.OrderField = v
	}

	if v, ok := c.GetQuery(NameOrderDirection); ok {
		opt.OrderDirection = strings.ToUpper(v)
	}

	var offset int
	if opt.PageNumber <= 1 {
		opt.PageNumber = 1
		offset = 0
	} else {
		offset = (opt.PageNumber - 1) * opt.PageSize
	}

	var totalCount int64
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}
	var orderStr []string
	if strings.Index(opt.OrderField, ",") > 0 {
		ofs := strings.Split(opt.OrderField, ",")
		for _, v := range ofs {
			orderStr = append(orderStr, fmt.Sprintf("%s %s", v, opt.OrderDirection))
		}
	} else {
		orderStr = append(orderStr, fmt.Sprintf("%s %s", opt.OrderField, opt.OrderDirection))
	}

	var ls []T
	if err := dbQuery.Order(strings.Join(orderStr, ",")).Limit(opt.PageSize).Offset(offset).Find(&ls).Error; err != nil {
		return nil, err
	}

	prs := &PaginationResult{
		PageSize:         opt.PageSize,
		Items:            ls,
		TotalCount:       totalCount,
		TotalPages:       int(math.Ceil(float64(totalCount) / float64(opt.PageSize))),
		PageNumber:       opt.PageNumber,
		CurrentPageCount: len(ls),
		OrderField:       opt.OrderField,
		OrderDirection:   opt.OrderDirection,
	}

	var liStr []string
	start := 1
	if prs.PageNumber > 2 {
		start = prs.PageNumber - 2
	}
	if start < 1 {
		start = 1
	}

	end := prs.PageNumber + 2
	if prs.PageNumber <= 2 {
		end = prs.PageNumber + 3
	}
	if end > prs.TotalPages {
		end = prs.TotalPages
	}

	for i := start; i <= end; i++ {
		if i == prs.PageNumber {
			liStr = append(liStr, fmt.Sprintf(`<li><span class="active">%d</span></li>`, i))
		} else {
			liStr = append(liStr, fmt.Sprintf(`<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: i}, false), i))
		}
	}

	paginationTemplate := []string{
		`<div class="pagination-wrapper">`,
		`<ul class="pagination">`,
		fmt.Sprintf(`<li><a href="%s">&laquo;</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: 1}, false)),
		func() string {
			i := prs.PageNumber - 1
			if i < 1 {
				i = 1
			}
			return fmt.Sprintf(`<li><a href="%s">&lang;</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: i}, false))
		}(),
		strings.Join(liStr, "\n"),
		func() string {
			i := prs.PageNumber + 1
			if i > prs.TotalPages {
				i = prs.TotalPages
			}
			return fmt.Sprintf(`<li><a href="%s">&rang;</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: i}, false))
		}(),
		fmt.Sprintf(`<li><a href="%s">&raquo;</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: prs.TotalPages}, false)),
		`</ul>`,
		`</div>`,
	}
	prs.PaginationHtml = template.HTML(strings.Join(paginationTemplate, "\n"))
	return prs, nil
}
