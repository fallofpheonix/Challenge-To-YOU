/*
 * CHRYSALIS MATRIX SOVEREIGN ARCHITECTURE
 *
 * FILE: fixedpoint.go
 * PATH: foundation/math/fixedpoint.go
 *
 * PURPOSE:
 * Canonical Fixed-Point arithmetic for Chrysalis Architecture.
 * Standardizes 10^6 scaling and saturating semantics to ensure bit-perfect
 * determinism across all system layers.
 */

package crysmath

import (
	"fmt"
	"math"
	"math/big"
)

const (
	Precision = 1000000 // 10^6 for 6 decimal places
)

var (
	bigMaxInt64  = big.NewInt(math.MaxInt64)
	bigMinInt64  = big.NewInt(math.MinInt64)
	bigPrecision = big.NewInt(Precision)
)

// FixedPoint represents a decimal number using a fixed-point integer representation.
type FixedPoint struct {
	V int64 `json:"v"`
}

// NewFixedPoint creates a FixedPoint from an integer numerator.
func NewFixedPoint(numerator int64) FixedPoint {
	return FixedPoint{V: numerator * Precision}
}

// NewFixedPointRaw creates a FixedPoint from a raw scaled value.
func NewFixedPointRaw(value int64) FixedPoint {
	return FixedPoint{V: value}
}

// Float64 converts to float64 for display/logging ONLY.
// NEVER USE FOR STATE TRANSITIONS OR CONSENSUS LOGIC.
func (f FixedPoint) Float64() float64 {
	return float64(f.V) / float64(Precision)
}

// String provides a human-readable representation.
func (f FixedPoint) String() string {
	return fmt.Sprintf("%.6f", f.Float64())
}

// SaturatingAdd adds two FixedPoint numbers with overflow protection.
func (f FixedPoint) SaturatingAdd(other FixedPoint) FixedPoint {
	a, b := f.V, other.V
	if a > 0 && b > math.MaxInt64-a {
		return FixedPoint{V: math.MaxInt64}
	}
	if a < 0 && b < math.MinInt64-a {
		return FixedPoint{V: math.MinInt64}
	}
	return FixedPoint{V: a + b}
}

// SaturatingSub subtracts two FixedPoint numbers with underflow protection.
func (f FixedPoint) SaturatingSub(other FixedPoint) FixedPoint {
	a, b := f.V, other.V
	if b > 0 && a < math.MinInt64+b {
		return FixedPoint{V: math.MinInt64}
	}
	if b < 0 && a > math.MaxInt64+b {
		return FixedPoint{V: math.MaxInt64}
	}
	return FixedPoint{V: a - b}
}

// Mul multiplies two FixedPoint numbers with saturating semantics.
// Rounding: Standardizes on truncation (round toward zero).
func (f FixedPoint) Mul(other FixedPoint) FixedPoint {
	x := new(big.Int).SetInt64(f.V)
	y := new(big.Int).SetInt64(other.V)
	p := new(big.Int).Mul(x, y)
	res := new(big.Int).Quo(p, bigPrecision)

	if res.Cmp(bigMaxInt64) > 0 {
		return FixedPoint{V: math.MaxInt64}
	}
	if res.Cmp(bigMinInt64) < 0 {
		return FixedPoint{V: math.MinInt64}
	}
	return FixedPoint{V: res.Int64()}
}

// Div divides two FixedPoint numbers with saturating semantics.
// Rounding: Standardizes on truncation (round toward zero).
func (f FixedPoint) Div(other FixedPoint) FixedPoint {
	if other.V == 0 {
		if f.V >= 0 {
			return FixedPoint{V: math.MaxInt64}
		}
		return FixedPoint{V: math.MinInt64}
	}

	x := new(big.Int).SetInt64(f.V)
	x.Mul(x, bigPrecision)
	y := new(big.Int).SetInt64(other.V)
	res := new(big.Int).Quo(x, y)

	if res.Cmp(bigMaxInt64) > 0 {
		return FixedPoint{V: math.MaxInt64}
	}
	if res.Cmp(bigMinInt64) < 0 {
		return FixedPoint{V: math.MinInt64}
	}
	return FixedPoint{V: res.Int64()}
}
