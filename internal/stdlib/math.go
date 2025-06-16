package stdlib

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	"github.com/dshills/alas/internal/runtime"
)

// registerMathFunctions registers all std.math builtin functions.
func (r *Registry) registerMathFunctions() {
	// Constants
	r.Register("math.PI", mathPI)
	r.Register("math.E", mathE)

	// Basic operations
	r.Register("math.abs", mathAbs)
	r.Register("math.min", mathMin)
	r.Register("math.max", mathMax)
	r.Register("math.pow", mathPow)
	r.Register("math.sqrt", mathSqrt)

	// Trigonometric functions
	r.Register("math.sin", mathSin)
	r.Register("math.cos", mathCos)
	r.Register("math.tan", mathTan)
	r.Register("math.asin", mathAsin)
	r.Register("math.acos", mathAcos)
	r.Register("math.atan", mathAtan)

	// Rounding functions
	r.Register("math.floor", mathFloor)
	r.Register("math.ceil", mathCeil)
	r.Register("math.round", mathRound)

	// Random functions
	r.Register("math.random", mathRandom)
	r.Register("math.randomInt", mathRandomInt)
}

// validateTwoFloatArgs validates that exactly 2 arguments are provided and converts them to floats.
func validateTwoFloatArgs(args []runtime.Value, funcName string) (float64, float64, error) {
	if len(args) != 2 {
		return 0, 0, fmt.Errorf("%s expects 2 arguments, got %d", funcName, len(args))
	}

	a, err := args[0].AsFloat()
	if err != nil {
		return 0, 0, fmt.Errorf("%s: first argument: %v", funcName, err)
	}

	b, err := args[1].AsFloat()
	if err != nil {
		return 0, 0, fmt.Errorf("%s: second argument: %v", funcName, err)
	}

	return a, b, nil
}

// mathPI implements math.PI builtin function (returns PI constant).
func mathPI(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 0 {
		return runtime.NewVoid(), fmt.Errorf("math.PI expects 0 arguments, got %d", len(args))
	}
	return runtime.NewFloat(math.Pi), nil
}

// mathE implements math.E builtin function (returns E constant).
func mathE(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 0 {
		return runtime.NewVoid(), fmt.Errorf("math.E expects 0 arguments, got %d", len(args))
	}
	return runtime.NewFloat(math.E), nil
}

// mathAbs implements math.abs builtin function.
func mathAbs(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.abs expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.abs: %v", err)
	}

	return runtime.NewFloat(math.Abs(val)), nil
}

// mathMin implements math.min builtin function.
func mathMin(args []runtime.Value) (runtime.Value, error) {
	a, b, err := validateTwoFloatArgs(args, "math.min")
	if err != nil {
		return runtime.NewVoid(), err
	}

	return runtime.NewFloat(math.Min(a, b)), nil
}

// mathMax implements math.max builtin function.
func mathMax(args []runtime.Value) (runtime.Value, error) {
	a, b, err := validateTwoFloatArgs(args, "math.max")
	if err != nil {
		return runtime.NewVoid(), err
	}

	return runtime.NewFloat(math.Max(a, b)), nil
}

// mathPow implements math.pow builtin function.
func mathPow(args []runtime.Value) (runtime.Value, error) {
	base, exp, err := validateTwoFloatArgs(args, "math.pow")
	if err != nil {
		return runtime.NewVoid(), err
	}

	return runtime.NewFloat(math.Pow(base, exp)), nil
}

// mathSqrt implements math.sqrt builtin function.
func mathSqrt(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.sqrt expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.sqrt: %v", err)
	}

	if val < 0 {
		return runtime.NewVoid(), fmt.Errorf("math.sqrt: square root of negative number")
	}

	return runtime.NewFloat(math.Sqrt(val)), nil
}

// mathSin implements math.sin builtin function.
func mathSin(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.sin expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.sin: %v", err)
	}

	return runtime.NewFloat(math.Sin(val)), nil
}

// mathCos implements math.cos builtin function.
func mathCos(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.cos expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.cos: %v", err)
	}

	return runtime.NewFloat(math.Cos(val)), nil
}

// mathTan implements math.tan builtin function.
func mathTan(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.tan expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.tan: %v", err)
	}

	return runtime.NewFloat(math.Tan(val)), nil
}

// mathAsin implements math.asin builtin function.
func mathAsin(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.asin expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.asin: %v", err)
	}

	if val < -1 || val > 1 {
		return runtime.NewVoid(), fmt.Errorf("math.asin: input out of range [-1, 1]")
	}

	return runtime.NewFloat(math.Asin(val)), nil
}

// mathAcos implements math.acos builtin function.
func mathAcos(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.acos expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.acos: %v", err)
	}

	if val < -1 || val > 1 {
		return runtime.NewVoid(), fmt.Errorf("math.acos: input out of range [-1, 1]")
	}

	return runtime.NewFloat(math.Acos(val)), nil
}

// mathAtan implements math.atan builtin function.
func mathAtan(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.atan expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.atan: %v", err)
	}

	return runtime.NewFloat(math.Atan(val)), nil
}

// mathFloor implements math.floor builtin function.
func mathFloor(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.floor expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.floor: %v", err)
	}

	return runtime.NewFloat(math.Floor(val)), nil
}

// mathCeil implements math.ceil builtin function.
func mathCeil(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.ceil expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.ceil: %v", err)
	}

	return runtime.NewFloat(math.Ceil(val)), nil
}

// mathRound implements math.round builtin function.
func mathRound(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("math.round expects 1 argument, got %d", len(args))
	}

	val, err := args[0].AsFloat()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.round: %v", err)
	}

	return runtime.NewFloat(math.Round(val)), nil
}

// mathRandom implements math.random builtin function.
func mathRandom(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 0 {
		return runtime.NewVoid(), fmt.Errorf("math.random expects 0 arguments, got %d", len(args))
	}

	// Generate cryptographically secure random float64 in range [0, 1)
	max := big.NewInt(1 << 53) // Use 53 bits for float64 precision
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.random: failed to generate random number: %v", err)
	}

	// Convert to float64 in range [0, 1)
	result := float64(n.Int64()) / float64(1<<53)
	return runtime.NewFloat(result), nil
}

// mathRandomInt implements math.randomInt builtin function.
func mathRandomInt(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("math.randomInt expects 2 arguments, got %d", len(args))
	}

	minVal, err := args[0].AsInt()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.randomInt: min: %v", err)
	}

	maxVal, err := args[1].AsInt()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.randomInt: max: %v", err)
	}

	if minVal > maxVal {
		return runtime.NewVoid(), fmt.Errorf("math.randomInt: min cannot be greater than max")
	}

	// Generate cryptographically secure random number in range [min, max] (inclusive)
	rangeSize := big.NewInt(maxVal - minVal + 1)
	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("math.randomInt: failed to generate random number: %v", err)
	}

	result := n.Int64() + minVal
	return runtime.NewInt(result), nil
}
