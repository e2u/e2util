package req

import (
	"testing"
)

func Test_ParsePayload(t *testing.T) {
	t.Run("RangePayload - header", func(t *testing.T) {
		if v, err := ParseRangePayload("products=8-11"); err != nil {
			t.Fatal(err)
		} else if v.Type != "products" || v.Start != 8 || v.End != 11 {
			t.Fatal(v)
		}
	})

	t.Run("RangePayload - query", func(t *testing.T) {
		if v, err := ParseRangePayload("[3,9]"); err != nil {
			t.Fatal(err)
		} else if v.Start != 3 || v.End != 9 {
			t.Fatal(v)
		}
	})

	t.Run("FilterPayload", func(t *testing.T) {

		t.Run("all", func(t *testing.T) {
			str := `{
  "stock_lt": 10,
  "stock_gt": 0,
  "sales": 0,
  "name_inc_any": [
    "cat",
    "dog"
  ],
  "price_eq":100.02,
  "ff_aa_neq":"na",
  "cc_neq_any":[10,20,11.04],
  "ddd_q":[true,false,true],
  "eee_lt":79.39,
  "fff_lte":12,
  "ggg_gt":101.01,
  "hhh_gte":89.01
  
}`
			_, err := ParseFilterPayload(str)
			if err != nil {
				t.Fatal(err)
			}
		})

		t.Run("only lt", func(t *testing.T) {
			str := `{"stock_lt":10}`
			rs, err := ParseFilterPayload(str)
			if err != nil {
				t.Fatal(err)
			}
			if rs[0].Field != "stock" {
				t.Fatal(rs[0].Field)
			}
			if rs[0].Operator != "_lt" {
				t.Fatal(rs[0].Operator)
			}
			if rs[0].Symbol != "<" {
				t.Fatal(rs[0].Symbol)
			}

		})

	})
}
