package bigdecimal

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBigDecimal_NewFromString(t *testing.T) {
	tests := []struct {
		value          string
		expectedBigInt string
		expectedScale  int64
	}{
		{"0.1", "1", 1},
		{"-0.1", "-1", 1},
		{"198.98765544", "19898765544", 8},
		{"0.00000093937698", "93937698", 14},
		{"98765587998098786876.0", "98765587998098786876", 0},
		{"98765000000", "98765", -6},
		{"-98765000000", "-98765", -6},
		{"98765000000.1", "987650000001", 1},
		{"-98765000000.2", "-987650000002", 1},

		// Positive rounding outside max scale (34)
		{"0.1234567890123456789012345678901234", "1234567890123456789012345678901234", 34},
		{"0.12345678901234567890123456789012344", "1234567890123456789012345678901234", 34},
		{"0.12345678901234567890123456789012345", "1234567890123456789012345678901235", 34},
		{"0.12345678901234567890123456789012346", "1234567890123456789012345678901235", 34},

		// Negative rounding outside max scale (34)
		{"-0.1234567890123456789012345678901234", "-1234567890123456789012345678901234", 34},
		{"-0.12345678901234567890123456789012344", "-12345678901234567890123456789012344", 35},
		{"-0.12345678901234567890123456789012345", "-12345678901234567890123456789012345", 35},
		{"-0.12345678901234567890123456789012346", "-12345678901234567890123456789012346", 35},

		// Normalize negative numbers have a bug where scale is actually MAX + 1
		{"-0.123456789012345678901234567890123424", "-12345678901234567890123456789012342", 35},
		{"-0.123456789012345678901234567890123425", "-12345678901234567890123456789012342", 35},
		{"-0.123456789012345678901234567890123426", "-12345678901234567890123456789012342", 35},

		// Showcasing rounding effects when max digits is split before/after dot
		{"12.123456789012345678901234567890124", "1212345678901234567890123456789012", 32},
		{"12.123456789012345678901234567890125", "1212345678901234567890123456789013", 32},
		{"12.123456789012345678901234567890126", "1212345678901234567890123456789013", 32},

		{"-12.1234567890123456789012345678901234", "-12123456789012345678901234567890123", 33},
		{"-12.1234567890123456789012345678901235", "-12123456789012345678901234567890123", 33},
		{"-12.1234567890123456789012345678901236", "-12123456789012345678901234567890123", 33},

		{"1234567890123.123456789012345678901834567890124", "1234567890123123456789012345678902", 21},
		{"-1234567890123.123456789012345678901894567890124", "-12345678901231234567890123456789018", 22},

		// Showcasing rounding effects when max digits is all before dot
		{"1234567890123456789012345678901234", "1234567890123456789012345678901234", 0},
		{"12345678901234567890123456789012344", "1234567890123456789012345678901234", -1},
		{"12345678901234567890123456789012345", "1234567890123456789012345678901235", -1},
		{"12345678901234567890123456789012346", "1234567890123456789012345678901235", -1},

		{"-12345678901234567890123456789012345", "-12345678901234567890123456789012345", 0},
		{"-123456789012345678901234567890123454", "-12345678901234567890123456789012345", -1},
		{"-123456789012345678901234567890123455", "-12345678901234567890123456789012345", -1},
		{"-123456789012345678901234567890123456", "-12345678901234567890123456789012345", -1},

		{"10000000000000000000000000000000000000000", "1", -40},
		{"100000000000000000000000000000000000000001", "1", -41},

		{"19999999999999999999999999999999994", "1999999999999999999999999999999999", -1},
		{"19999999999999999999999999999999995", "2", -34},
		{"19999999999999999999999999999999985", "1999999999999999999999999999999999", -1},

		{"1999999999999999999999999999999999", "1999999999999999999999999999999999", 0},
		{"199999999999999999999999999999999", "199999999999999999999999999999999", 0},
		{"19999999999999999999999999999999999", "2", -34},
		{"199999999999999999999999999999999999999999", "2", -41},

		{"1444444444444444444444444444444444", "1444444444444444444444444444444444", 0},
		{"14444444444444444444444444444444444", "1444444444444444444444444444444444", -1},
		{"144444444444444444444444444444444444", "1444444444444444444444444444444444", -2},

		{"1555555555555555555555555555555555", "1555555555555555555555555555555555", 0},
		{"15555555555555555555555555555555555", "1555555555555555555555555555555556", -1},
		{"155555555555555555555555555555555555", "1555555555555555555555555555555556", -2},
		{"0.00000000000000000000000000000000", "0", 0},
		{"0.10000000000000000000000000000000", "1", 1},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			// After many tests in `graph-node`, the rounding after precision goes over is to
			// go Toward Positive Infinity (rounding up if number is positive, truncating if number is negative)
			//
			// See https://en.wikipedia.org/wiki/IEEE_754#Directed_roundings (toward +∞)

			expectedBigInt, ok := (&big.Int{}).SetString(tt.expectedBigInt, 10)
			require.True(t, ok)

			actual, err := NewFromString(tt.value)
			require.NoError(t, err)
			require.NotNil(t, actual.Int)

			msg := []any{
				"For %s [BigInt (expected %s, actual %s), Scale (expected %d, actual %d)]",
				tt.value,
				tt.expectedBigInt, actual.Int,
				tt.expectedScale, actual.Scale,
			}

			assert.True(t, expectedBigInt.Cmp(actual.Int) == 0, msg...)
			assert.Equal(t, tt.expectedScale, actual.Scale, msg...)
		})
	}

}

func TestBigDecimal_Cmp(t *testing.T) {
	type args struct {
		left  string
		right string
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{"zero", args{"0", "0"}, 0},
		{"small_positive", args{"12.34", "1.234"}, 1},
		{"small_positive_inverted", args{"1.234", "12.34"}, -1},
		{"small_negative", args{"12.34", "-1.234"}, 1},
		{"small_negative_inverted", args{"-1.234", "12.34"}, -1},
		{"equal_decimals", args{"1.23", "1.23"}, 0},
		{"equal_negative", args{"-1.23", "-1.23"}, 0},
		{"two_negative", args{"-1.23", "-4.23"}, 1},
		{"really_big_+_really_small", args{"1234e6", "1234e-6"}, 1},
		{"really_small_+_really_big", args{"1234e-6", "1234e6"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := MustNewFromString(tt.args.left)
			y := MustNewFromString(tt.args.right)

			assert.Equal(t, tt.want, x.Cmp(y), "x.{Int, Scale}: {%s, %d}, y.{Int, Scale}: {%s, %d}", x.Int, x.Scale, y.Int, y.Scale)
		})
	}
}

func TestBigDecimal_Add(t *testing.T) {
	type args struct {
		left  string
		right string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"zero", args{"0", "0"}, "0"},
		{"zero_with_decs", args{"0.000000000000000000000000000000000000000", "0.000000000000000000000000000000000000000"}, "0"},
		{"small_positive", args{"12.34", "1.234"}, "13.574"},
		{"small_positive_inverted", args{"1.234", "12.34"}, "13.574"},
		{"small_negative", args{"12.34", "-1.234"}, "11.106"},
		{"small_negative_inverted", args{"-1.234", "12.34"}, "11.106"},
		{"equal_decimals", args{"1.23", "1.23"}, "2.46"},
		{"really_big_+_really_small", args{"1234e6", "1234e-6"}, "1234000000.001234"},
		{"really_small_+_really_big", args{"1234e-6", "1234e6"}, "1234000000.001234"},
		{"no_meaning_decimals_+_1", args{"18446744073709551616.0", "1"}, "18446744073709551617"},
		{"always_expanded", args{"184467440737e6", "0"}, "184467440737000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := New()
			x := MustNewFromString(tt.args.left)
			y := MustNewFromString(tt.args.right)

			assert.Equal(t, tt.want, z.Add(x, y).String(), "x.{Int, Scale}: {%s, %d}, y.{Int, Scale}: {%s, %d} equals z.{Int, Scale}: {%s, %d}", x.Int, x.Scale, y.Int, y.Scale, z.Int, z.Scale)
		})
	}
}

func TestBigDecimal_String(t *testing.T) {
	type args struct {
		value int
		scale int
	}

	tests := []struct {
		want string
		args args
	}{
		{"1", args{1, 0}},
		{"0.1", args{1, 1}},
		{"0.01", args{1, 2}},
		{"100", args{1, -2}},
		{"-1", args{-1, 0}},
		{"-0.1", args{-1, 1}},
		{"-0.01", args{-1, 2}},
		{"13.574", args{13574, 3}},
		{"132.4", args{132400000, 6}},
		{"0", args{0, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			z := BigDecimal{Int: big.NewInt(int64(tt.args.value)), Scale: int64(tt.args.scale)}
			assert.Equal(t, tt.want, z.String())

			// Round-trip should work
			z = *MustNewFromString(tt.want)
			assert.Equal(t, tt.want, z.String())
		})
	}
}
