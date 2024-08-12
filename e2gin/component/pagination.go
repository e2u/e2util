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

const (
	NamePageSize       = "_ps"
	NamePageNumber     = "_pn"
	NameOrderField     = "_of"
	NameOrderDirection = "_od"

	ValueOrderByAsc  = "ASC"
	ValueOrderByDesc = "DESC"

	DefaultOrderField = "pkaid"

	JSONItemsKey = "items"
)

type PaginationResult struct {
	Items          any    `json:"items,omitempty"`
	Total          int64  `json:"total"`
	Pages          int    `json:"pages"`
	PrePage        int    `json:"pre_page"`
	Offset         int    `json:"offset"`
	Page           int    `json:"page"`
	CurrentPages   int    `json:"current_pages"`
	Html           any    `json:"html,omitempty"`
	OrderField     string `json:"order_field,omitempty"`
	OrderDirection string `json:"order_direction,omitempty"`
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
	PrePage        int
	Page           int
	Offset         int    //  offset priority highter than page
	OrderField     string // _of=pkaid | _of=size | _of=created_at
	OrderDirection string // _or=ASC | _or=DESC
	DisableHtmlBar bool
}

func PaginationList[T any](c *gin.Context, model T, dbQuery *gorm.DB, opts ...*PaginationOption) (*PaginationResult, error) {
	dbQuery = dbQuery.Model(model)
	var opt *PaginationOption

	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = &PaginationOption{
			PrePage:        10,
			Page:           1,
			OrderDirection: ValueOrderByDesc,
			OrderField:     DefaultOrderField,
		}
	}

	if opt.OrderField == "" {
		opt.OrderField = DefaultOrderField
	}

	if opt.OrderDirection == "" {
		opt.OrderDirection = ValueOrderByDesc
	}

	if v, ok := c.GetQuery(NamePageSize); ok {
		opt.PrePage = e2strconv.MustParseInt(v)
	}

	if v, ok := c.GetQuery(NamePageNumber); ok {
		opt.Page = e2strconv.MustParseInt(v)
	}

	if v, ok := c.GetQuery(NameOrderField); ok {
		opt.OrderField = v
	}

	if v, ok := c.GetQuery(NameOrderDirection); ok {
		opt.OrderDirection = strings.ToUpper(v)
	}

	if opt.PrePage <= 0 {
		opt.PrePage = 10
	}
	if opt.Page <= 0 {
		opt.Page = 1
	}

	var offset int
	if opt.Page <= 1 {
		opt.Page = 1
		offset = 0
	} else {
		offset = (opt.Page - 1) * opt.PrePage
	}

	if opt.Offset > 0 {
		offset = opt.Offset
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
	if err := dbQuery.Order(strings.Join(orderStr, ",")).Limit(opt.PrePage).Offset(offset).Find(&ls).Error; err != nil {
		return nil, err
	}

	prs := &PaginationResult{
		PrePage:        opt.PrePage,
		Items:          ls,
		Total:          totalCount,
		Pages:          int(math.Ceil(float64(totalCount) / float64(opt.PrePage))),
		Page:           opt.Page,
		Offset:         offset,
		CurrentPages:   len(ls),
		OrderField:     opt.OrderField,
		OrderDirection: opt.OrderDirection,
	}
	if totalCount <= 0 {
		return prs, nil
	}

	if opt.DisableHtmlBar {
		return prs, nil
	}

	var liStr []string
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
			liStr = append(liStr, fmt.Sprintf(`<li><span class="active">%d</span></li>`, i))
		} else {
			liStr = append(liStr, fmt.Sprintf(`<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: i}, false), i))
		}
	}

	paginationTemplate := []string{
		`<div class="pagination-wrapper">`,
		`<ul class="pagination">`,
		func() string {
			if prs.Page > 3 {
				return fmt.Sprintf(`<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: 1}, false), 1)
			}
			return ""
		}(),
		strings.Join(liStr, "\n"),
		func() string {
			if prs.Page < prs.Pages {
				return fmt.Sprintf(`<li><a href="%s">%d</a></li>`, BuildQueryUri(c, map[string]any{NamePageNumber: prs.Pages}, false), prs.Pages)
			}
			return ""
		}(),
		`</ul>`,
		`</div>`,
	}
	prs.Html = template.HTML(strings.Join(paginationTemplate, "")) // #nosec G203
	return prs, nil
}
