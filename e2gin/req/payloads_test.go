package req

import (
	"reflect"
	"testing"
)

func Test_ParsePayload(t *testing.T) {
	// RangePayload tests
	t.Run("RangePayload - header", func(t *testing.T) {
		if v, err := ParseRangePayload("products=8-11"); err != nil {
			t.Fatal(err)
		} else if v.Type != "products" || v.Start != 8 || v.End != 11 {
			t.Fatalf("unexpected RangePayload: %+v", v)
		}
	})

	t.Run("RangePayload - query", func(t *testing.T) {
		if v, err := ParseRangePayload("[3,9]"); err != nil {
			t.Fatal(err)
		} else if v.Start != 3 || v.End != 9 {
			t.Fatalf("unexpected RangePayload: %+v", v)
		}
	})

	t.Run("RangePayload - invalid", func(t *testing.T) {
		if _, err := ParseRangePayload("not-json-and-not-range"); err == nil {
			t.Fatal("expected error for invalid range payload")
		}
	})

	// FilterPayload tests
	t.Run("FilterPayload - all operators", func(t *testing.T) {
		str := `{
  "stock@lt": 10,
  "stock@gt": 0,
  "sales": 0,
  "name@inc_any": [
    "cat",
    "dog"
  ],
  "price@eq":100.02,
  "ff_aa@neq":"na",
  "cc@eq_any":[10,20,11.04],
  "ddd@q":[true,false,true],
  "eee@lt":79.39,
  "fff@lte":12,
  "ggg@gt":101.01,
  "hhh@gte":89.01,
  "mmm@like":"ab",
  "nnn@nlike":"cd",
  "is1@is_null":true,
  "is2@is_not_null":true,
  "empty@is_empty":true
}`
		rs, err := ParseFilterPayload(str)
		if err != nil {
			t.Fatal(err)
		}

		// Convert slice to lookup by field+operator for order-independent assertions
		lookup := map[string]Filter{}
		for _, f := range rs {
			lookup[f.Field+"|"+f.Operator] = f
		}

		cases := []struct{
			key string
			field string
			op string
			symbol string
		}{
			{"stock@lt","stock","@lt","<"},
			{"stock@gt","stock","@gt",">"},
			{"sales","sales","@eq","="},
			{"name@inc_any","name","@inc_any","IN"},
			{"price@eq","price","@eq","="},
			{"ff_aa@neq","ff_aa","@neq","!="},
			{"cc@eq_any","cc","@eq_any","IN"},
			{"ddd@q","ddd","@q","IN"},
			{"eee@lt","eee","@lt","<"},
			{"fff@lte","fff","@lte","<= "},
			{"ggg@gt","ggg","@gt",">"},
			{"hhh@gte","hhh","@gte",">="},
			{"mmm@like","mmm","@like","LIKE"},
			{"nnn@nlike","nnn","@nlike","NOT LIKE"},
			{"is1@is_null","is1","@is_null","IS NULL"},
			{"is2@is_not_null","is2","@is_not_null","IS NOT NULL"},
			{"empty@is_empty","empty","@is_empty","IS EMPTY"},
		}
		for _, c := range cases {
			f, ok := lookup[c.field+"|"+c.op]
			if !ok {
				t.Fatalf("missing filter for %s (%s)", c.field, c.op)
			}
			if f.Symbol != c.symbol {
				t.Fatalf("%s symbol got %q want %q", c.key, f.Symbol, c.symbol)
			}
		}
	})

	t.Run("FilterPayload - default operator and unknown suffix", func(t *testing.T) {
		str := `{"a":1, "b@unknown":2, "stock@lt":5}`
		rs, err := ParseFilterPayload(str)
		if err != nil {
			t.Fatal(err)
		}
		lookup := map[string]Filter{}
		for _, f := range rs {
			lookup[f.Field+"|"+f.Operator] = f
		}
		if f, ok := lookup["a|@eq"]; !ok || f.Symbol != "=" {
			t.Fatalf("default operator for 'a' not applied correctly: %+v", f)
		}
		// unknown suffix should not be stripped; field stays as original key, operator defaults to @eq
		if f, ok := lookup["b@unknown|@eq"]; !ok || f.Field != "b@unknown" || f.Symbol != "=" {
			t.Fatalf("unknown operator handling incorrect: %+v", f)
		}
		if f, ok := lookup["stock|@lt"]; !ok || f.Symbol != "<" {
			t.Fatalf("lt operator not parsed: %+v", f)
		}
	})

	// SortPayload tests
	t.Run("SortPayload - valid", func(t *testing.T) {
		sp, err := ParseSortPayload(`["name","DESC"]`)
		if err != nil {
			t.Fatal(err)
		}
		if sp.Field != "name" || sp.Order != "DESC" {
			t.Fatalf("unexpected SortPayload: %+v", sp)
		}
	})

	t.Run("SortPayload - invalid json returns default with error", func(t *testing.T) {
		sp, err := ParseSortPayload("not-json")
		if err == nil {
			t.Fatal("expected error for invalid sort payload")
		}
		if sp.Field != "id" || sp.Order != "ASC" {
			t.Fatalf("expected default sort, got: %+v", sp)
		}
	})

	t.Run("SortPayload - wrong length returns default without error", func(t *testing.T) {
		sp, err := ParseSortPayload(`["onlyfield"]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sp.Field != "id" || sp.Order != "ASC" {
			t.Fatalf("expected default sort, got: %+v", sp)
		}
	})

	// PaginationPayload tests
	t.Run("PaginationPayload - valid", func(t *testing.T) {
		pp, err := ParsePaginationPayload(`{"page":2,"perPage":50}`)
		if err != nil {
			t.Fatal(err)
		}
		if pp.Page != 2 || pp.PrePage != 50 {
			t.Fatalf("unexpected PaginationPayload: %+v", pp)
		}
	})

	t.Run("PaginationPayload - partial and defaults", func(t *testing.T) {
		pp, err := ParsePaginationPayload(`{}`)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(pp, PaginationPayload{Page:1, PrePage:10}) {
			t.Fatalf("expected defaults, got: %+v", pp)
		}
		pp2, err := ParsePaginationPayload(`{"page":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if pp2.Page != 3 || pp2.PrePage != 10 {
			t.Fatalf("expected default perPage, got: %+v", pp2)
		}
	})

	t.Run("PaginationPayload - invalid json returns default with error", func(t *testing.T) {
		pp, err := ParsePaginationPayload("bad-json")
		if err == nil {
			t.Fatal("expected error for invalid pagination payload")
		}
		if pp.Page != 1 || pp.PrePage != 10 {
			t.Fatalf("expected defaults on error, got: %+v", pp)
		}
	})
}
