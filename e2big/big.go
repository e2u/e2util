package e2big

import (
	"github.com/shopspring/decimal"
)

// 大数计算封装

var (
	Zero = decimal.NewFromFloat(0.0)
)

// FloatAdd 浮点大数加
func FloatAdd(x, y float64) decimal.Decimal {
	xx, yy := decimal.NewFromFloat(x), decimal.NewFromFloat(y)
	return xx.Add(yy)
}

// FloatAddN 支持2个以上的数从左到右相加
func FloatAddN(x float64, ys ...float64) decimal.Decimal {
	xx := decimal.NewFromFloat(x)
	for _, y := range ys {
		xx = xx.Add(decimal.NewFromFloat(y))
	}
	return xx
}

// FloatSub 浮点大数减
func FloatSub(x, y float64) decimal.Decimal {
	xx, yy := decimal.NewFromFloat(x), decimal.NewFromFloat(y)
	return xx.Sub(yy)
}

// FloatSubN 支持2个以上的数从左到右相减
func FloatSubN(x float64, ys ...float64) decimal.Decimal {
	xx := decimal.NewFromFloat(x)
	for _, y := range ys {
		xx = xx.Sub(decimal.NewFromFloat(y))
	}
	return xx
}

// FloatMul 浮点大数乘
func FloatMul(x, y float64) decimal.Decimal {
	xx, yy := decimal.NewFromFloat(x), decimal.NewFromFloat(y)
	return xx.Mul(yy)
}

// FloatMulN 支持2个以上的数从左到右相乘
func FloatMulN(x float64, ys ...float64) decimal.Decimal {
	xx := decimal.NewFromFloat(x)
	for _, y := range ys {
		xx = xx.Mul(decimal.NewFromFloat(y))
	}
	return xx
}

// FloatDiv 浮点大数除
func FloatDiv(x, y float64) decimal.Decimal {
	xx, yy := decimal.NewFromFloat(x), decimal.NewFromFloat(y)
	return xx.Div(yy)
}

// FloatDivN 支持2个以上的数从左到右相除
func FloatDivN(x float64, ys ...float64) decimal.Decimal {
	xx := decimal.NewFromFloat(x)
	for _, y := range ys {
		xx = xx.Div(decimal.NewFromFloat(y))
	}
	return xx
}

// FloatMod 浮点大数模
func FloatMod(x, y float64) decimal.Decimal {
	xx, yy := decimal.NewFromFloat(x), decimal.NewFromFloat(y)
	return xx.Mod(yy)
}
