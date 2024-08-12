package req

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/e2u/e2util/e2json"
	"github.com/e2u/e2util/e2map"
	"github.com/e2u/e2util/e2regexp"
	"golang.org/x/exp/maps"
)

//https://marmelab.com/react-admin/FilteringTutorial.html

/**
dataProvider.getList('posts', {
    filter: { commentable: true, q: 'lorem ' },
    pagination: { page: 1, perPage: 10 },
    sort: { field: 'published_at', order: 'DESC' },
});
*/

type SortPayload struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

func ParseSortPayload(s string) (SortPayload, error) {
	var sa []string
	err := e2json.MustFromJSONString(s, &sa)
	if err != nil || len(sa) != 2 {
		return SortPayload{
			Field: "id",
			Order: "ASC",
		}, err
	}
	return SortPayload{
		Field: sa[0],
		Order: sa[1],
	}, nil
}

type PaginationPayload struct {
	Page    int `json:"page"`
	PrePage int `json:"perPage"`
}

func ParsePaginationPayload(s string) (PaginationPayload, error) {
	r := PaginationPayload{Page: 1, PrePage: 10}
	err := e2json.MustFromJSONString(s, &r)
	return r, err
}

// Filter https://github.com/marmelab/FakeRest?tab=readme-ov-file#supported-filters
type Filter struct {
	Field    string `json:"field"`
	Value    any    `json:"value"`
	Operator string `json:"operator"`
	Symbol   string `json:"symbol"`
}

var operatorSymbol = e2map.Map{
	"_eq":      "=",      // check for equality on simple values: filter={"price_eq":20} // return books where the price is equal to 20
	"_neq":     "!=",     // check for inequality on simple values, filter={"price_neq":20} // return books where the price is not equal to 20
	"_eq_any":  "IN",     // check for equality on any passed values, filter={"price_eq_any":[20, 30]} // return books where the price is equal to 20 or 30
	"_neq_any": "NOT IN", // check for inequality on any passed values, filter={"price_neq_any":[20, 30]} // return books where the price is not equal to 20 nor 30
	"_inc_any": "IN",     // check for items that include any of the passed values, filter={"authors_inc_any":['William Gibson', 'Pat Cadigan']} // return books where authors include either 'William Gibson' or 'Pat Cadigan' or both
	"_q":       "IN",     // check for items that contain the provided text, filter={"author_q":['Gibson']} // return books where the author includes 'Gibson' not considering the other fields
	"_lt":      "<",      // check for items that have a value lower than the provided value, filter={"price_lte":100} // return books that have a price lower than 100
	"_lte":     "<= ",    // check for items that have a value lower than or equal to the provided value, filter={"price_lte":100} // return books that have a price lower or equal to 100
	"_gt":      ">",      // check for items that have a value greater than the provided value, filter={"price_gte":100} // return books that have a price greater than 100
	"_gte":     ">=",     // check for items that have a value greater than or equal to the provided value, filter={"price_gte":100} // return books that have a price greater or equal to 100
}

func ParseFilterPayload(s string) ([]Filter, error) {
	r := make(e2map.Map)
	err := e2json.MustFromJSONString(s, &r)
	if err != nil {
		return nil, err
	}
	sortedOperatorKeys := maps.Keys(operatorSymbol)
	slices.SortFunc(sortedOperatorKeys, func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})

	var rs []Filter

	r.Range(func(key string, value any) {
		tr := Filter{Value: value}
		for _, operator := range sortedOperatorKeys {
			if strings.HasSuffix(key, operator) {
				tr.Operator = operator
				tr.Field = strings.TrimSuffix(key, operator)
				tr.Symbol, _ = operatorSymbol.DefaultString(operator, "")
				continue
			}
		}
		if tr.Operator == "" {
			tr.Operator = "eq"
			tr.Field = key
			tr.Symbol = "="
		}
		rs = append(rs, tr)
	})

	fmt.Printf("%#v", rs)
	return rs, nil
}

type RangePayload struct {
	Type  string
	Start int
	End   int
}

func ParseRangePayload(s string) (RangePayload, error) {
	r := RangePayload{}
	regex := regexp.MustCompile(`^(?P<type>.+)=(?P<start>\d+)-(?P<end>\d+)$`)

	if mv, ok := e2regexp.NamedFindStringSubmatch(s, regex); ok {
		if v, vok := mv.DefaultString("type", ""); vok {
			r.Type = v
		}
		if v, vok := mv.DefaultInt("start", 0); vok {
			r.Start = v
		}

		if v, vok := mv.DefaultInt("end", 9); vok {
			r.End = v
		}
		return r, nil
	}

	var qa []int
	if err := e2json.MustFromJSONString(s, &qa); err != nil {
		return r, err
	}
	r.Start = qa[0]
	r.End = qa[1]

	return r, nil
}
