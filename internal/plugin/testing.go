package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dshills/alas/internal/runtime"
)

// TestSuite represents a plugin test suite
type TestSuite struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Plugin      string     `json:"plugin"`
	Tests       []TestCase `json:"tests"`
	Setup       []TestStep `json:"setup,omitempty"`
	Teardown    []TestStep `json:"teardown,omitempty"`
}

// TestCase represents a single test case
type TestCase struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Function    string                 `json:"function"`
	Args        []interface{}          `json:"args"`
	Expected    interface{}            `json:"expected"`
	ExpectedErr string                 `json:"expected_error,omitempty"`
	Timeout     string                 `json:"timeout,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Setup       []TestStep             `json:"setup,omitempty"`
	Cleanup     []TestStep             `json:"cleanup,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// TestStep represents a setup or teardown step
type TestStep struct {
	Type     string                 `json:"type"`
	Function string                 `json:"function,omitempty"`
	Args     []interface{}          `json:"args,omitempty"`
	Assign   string                 `json:"assign,omitempty"`
	Value    interface{}            `json:"value,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}

// TestResult represents the result of running a test
type TestResult struct {
	Name      string        `json:"name"`
	Passed    bool          `json:"passed"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Output    interface{}   `json:"output,omitempty"`
	Logs      []string      `json:"logs,omitempty"`
}

// TestRunner executes plugin tests
type TestRunner struct {
	registry *Registry
	manager  *SecurityManager
	results  []TestResult
	verbose  bool
}

// NewTestRunner creates a new test runner
func NewTestRunner(registry *Registry, manager *SecurityManager) *TestRunner {
	return &TestRunner{
		registry: registry,
		manager:  manager,
		results:  make([]TestResult, 0),
		verbose:  false,
	}
}

// SetVerbose enables verbose output
func (tr *TestRunner) SetVerbose(verbose bool) {
	tr.verbose = verbose
}

// LoadTestSuite loads a test suite from a file
func (tr *TestRunner) LoadTestSuite(path string) (*TestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite file: %w", err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse test suite JSON: %w", err)
	}

	return &suite, nil
}

// RunTestSuite executes all tests in a test suite
func (tr *TestRunner) RunTestSuite(suite *TestSuite) error {
	if tr.verbose {
		fmt.Printf("Running test suite: %s\n", suite.Name)
		fmt.Printf("Description: %s\n", suite.Description)
		fmt.Printf("Plugin: %s\n", suite.Plugin)
	}

	// Ensure plugin is loaded
	plugin, exists := tr.registry.Get(suite.Plugin)
	if !exists {
		return fmt.Errorf("plugin %s not found", suite.Plugin)
	}

	if plugin.State != StateLoaded {
		if err := tr.registry.Load(suite.Plugin); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", suite.Plugin, err)
		}
	}

	// Run setup steps
	if len(suite.Setup) > 0 {
		if err := tr.runSteps(suite.Setup, "setup"); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}
	}

	// Run individual tests
	for _, testCase := range suite.Tests {
		result := tr.runTestCase(plugin, &testCase)
		tr.results = append(tr.results, result)

		if tr.verbose {
			if result.Passed {
				fmt.Printf("  ✓ %s (%v)\n", result.Name, result.Duration)
			} else {
				fmt.Printf("  ✗ %s (%v): %s\n", result.Name, result.Duration, result.Error)
			}
		}
	}

	// Run teardown steps
	if len(suite.Teardown) > 0 {
		if err := tr.runSteps(suite.Teardown, "teardown"); err != nil {
			return fmt.Errorf("teardown failed: %w", err)
		}
	}

	return nil
}

// runTestCase executes a single test case
func (tr *TestRunner) runTestCase(plugin *Plugin, testCase *TestCase) TestResult {
	start := time.Now()
	result := TestResult{
		Name:   testCase.Name,
		Passed: false,
	}

	// Run test setup
	if len(testCase.Setup) > 0 {
		if err := tr.runSteps(testCase.Setup, "test setup"); err != nil {
			result.Duration = time.Since(start)
			result.Error = fmt.Sprintf("test setup failed: %v", err)
			return result
		}
	}

	// Parse timeout
	timeout := 5 * time.Second // default
	if testCase.Timeout != "" {
		if parsed, err := time.ParseDuration(testCase.Timeout); err == nil {
			timeout = parsed
		}
	}

	// Execute test with timeout
	done := make(chan struct{})
	var output runtime.Value
	var err error

	go func() {
		defer close(done)
		// Convert args to runtime values
		args := make([]runtime.Value, len(testCase.Args))
		for i, arg := range testCase.Args {
			args[i] = tr.toRuntimeValue(arg)
		}

		// Call plugin function
		if plugin.Loader != nil {
			output, err = plugin.Loader.Call(plugin, testCase.Function, args)
		} else {
			err = fmt.Errorf("plugin not loaded")
		}
	}()

	select {
	case <-done:
		// Test completed
	case <-time.After(timeout):
		err = fmt.Errorf("test timed out after %v", timeout)
	}

	result.Duration = time.Since(start)

	// Check results
	if err != nil {
		if testCase.ExpectedErr != "" {
			// Expected an error
			if err.Error() == testCase.ExpectedErr {
				result.Passed = true
			} else {
				result.Error = fmt.Sprintf("expected error '%s', got '%s'", testCase.ExpectedErr, err.Error())
			}
		} else {
			result.Error = err.Error()
		}
	} else {
		if testCase.ExpectedErr != "" {
			// Expected an error but got none
			result.Error = fmt.Sprintf("expected error '%s', but test passed", testCase.ExpectedErr)
		} else {
			// Compare output with expected
			if tr.compareValues(output, testCase.Expected) {
				result.Passed = true
				result.Output = tr.fromRuntimeValue(output)
			} else {
				result.Error = fmt.Sprintf("expected %v, got %v", testCase.Expected, tr.fromRuntimeValue(output))
			}
		}
	}

	// Run test cleanup
	if len(testCase.Cleanup) > 0 {
		if cleanupErr := tr.runSteps(testCase.Cleanup, "test cleanup"); cleanupErr != nil {
			if result.Passed {
				result.Error = fmt.Sprintf("test passed but cleanup failed: %v", cleanupErr)
				result.Passed = false
			} else {
				result.Error = fmt.Sprintf("%s; cleanup also failed: %v", result.Error, cleanupErr)
			}
		}
	}

	return result
}

// runSteps executes a sequence of test steps
func (tr *TestRunner) runSteps(steps []TestStep, context string) error {
	for i, step := range steps {
		if err := tr.runStep(&step); err != nil {
			return fmt.Errorf("%s step %d failed: %w", context, i+1, err)
		}
	}
	return nil
}

// runStep executes a single test step
func (tr *TestRunner) runStep(step *TestStep) error {
	switch step.Type {
	case "call":
		// Function call step
		return fmt.Errorf("function call steps not yet implemented")
	case "assign":
		// Variable assignment step
		return fmt.Errorf("assignment steps not yet implemented")
	case "delay":
		// Delay step
		if durationStr, ok := step.Meta["duration"].(string); ok {
			if duration, err := time.ParseDuration(durationStr); err == nil {
				time.Sleep(duration)
				return nil
			}
		}
		return fmt.Errorf("invalid delay step")
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// toRuntimeValue converts a test value to a runtime value
func (tr *TestRunner) toRuntimeValue(value interface{}) runtime.Value {
	switch v := value.(type) {
	case nil:
		return runtime.NewVoid()
	case bool:
		return runtime.NewBool(v)
	case int:
		return runtime.NewInt(int64(v))
	case int64:
		return runtime.NewInt(v)
	case float64:
		return runtime.NewFloat(v)
	case string:
		return runtime.NewString(v)
	case []interface{}:
		elements := make([]runtime.Value, len(v))
		for i, elem := range v {
			elements[i] = tr.toRuntimeValue(elem)
		}
		return runtime.NewArray(elements)
	case map[string]interface{}:
		pairs := make(map[string]runtime.Value)
		for k, val := range v {
			pairs[k] = tr.toRuntimeValue(val)
		}
		return runtime.NewMap(pairs)
	default:
		return runtime.NewVoid()
	}
}

// fromRuntimeValue converts a runtime value to a test value
func (tr *TestRunner) fromRuntimeValue(value runtime.Value) interface{} {
	switch value.Type {
	case runtime.ValueTypeVoid:
		return nil
	case runtime.ValueTypeBool:
		if b, err := value.AsBool(); err == nil {
			return b
		}
		return false
	case runtime.ValueTypeInt:
		if i, err := value.AsInt(); err == nil {
			return i
		}
		return 0
	case runtime.ValueTypeFloat:
		if f, err := value.AsFloat(); err == nil {
			return f
		}
		return 0.0
	case runtime.ValueTypeString:
		if s, err := value.AsString(); err == nil {
			return s
		}
		return ""
	case runtime.ValueTypeArray:
		if arr, err := value.AsArray(); err == nil {
			result := make([]interface{}, len(arr))
			for i, elem := range arr {
				result[i] = tr.fromRuntimeValue(elem)
			}
			return result
		}
		return []interface{}{}
	case runtime.ValueTypeMap:
		if m, err := value.AsMap(); err == nil {
			result := make(map[string]interface{})
			for k, val := range m {
				result[k] = tr.fromRuntimeValue(val)
			}
			return result
		}
		return map[string]interface{}{}
	default:
		return nil
	}
}

// compareValues compares a runtime value with an expected test value
func (tr *TestRunner) compareValues(actual runtime.Value, expected interface{}) bool {
	expectedValue := tr.toRuntimeValue(expected)
	return tr.valuesEqual(actual, expectedValue)
}

// valuesEqual checks if two runtime values are equal
func (tr *TestRunner) valuesEqual(a, b runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case runtime.ValueTypeVoid:
		return true
	case runtime.ValueTypeBool:
		aBool, aErr := a.AsBool()
		bBool, bErr := b.AsBool()
		return aErr == nil && bErr == nil && aBool == bBool
	case runtime.ValueTypeInt:
		aInt, aErr := a.AsInt()
		bInt, bErr := b.AsInt()
		return aErr == nil && bErr == nil && aInt == bInt
	case runtime.ValueTypeFloat:
		aFloat, aErr := a.AsFloat()
		bFloat, bErr := b.AsFloat()
		return aErr == nil && bErr == nil && aFloat == bFloat
	case runtime.ValueTypeString:
		aStr, aErr := a.AsString()
		bStr, bErr := b.AsString()
		return aErr == nil && bErr == nil && aStr == bStr
	case runtime.ValueTypeArray:
		aArr, aErr := a.AsArray()
		bArr, bErr := b.AsArray()
		if aErr != nil || bErr != nil || len(aArr) != len(bArr) {
			return false
		}
		for i := range aArr {
			if !tr.valuesEqual(aArr[i], bArr[i]) {
				return false
			}
		}
		return true
	case runtime.ValueTypeMap:
		aMap, aErr := a.AsMap()
		bMap, bErr := b.AsMap()
		if aErr != nil || bErr != nil || len(aMap) != len(bMap) {
			return false
		}
		for k, aVal := range aMap {
			bVal, exists := bMap[k]
			if !exists || !tr.valuesEqual(aVal, bVal) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// GetResults returns the test results
func (tr *TestRunner) GetResults() []TestResult {
	return tr.results
}

// GetSummary returns a test summary
func (tr *TestRunner) GetSummary() TestSummary {
	summary := TestSummary{
		Total: len(tr.results),
	}

	for _, result := range tr.results {
		if result.Passed {
			summary.Passed++
		} else {
			summary.Failed++
		}
		summary.Duration += result.Duration
	}

	return summary
}

// TestSummary represents a summary of test results
type TestSummary struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Duration time.Duration `json:"duration"`
}

// RunGoTest integrates with Go's testing framework
func RunGoTest(t *testing.T, testSuitePath string, registry *Registry, manager *SecurityManager) {
	runner := NewTestRunner(registry, manager)
	runner.SetVerbose(testing.Verbose())

	suite, err := runner.LoadTestSuite(testSuitePath)
	if err != nil {
		t.Fatalf("Failed to load test suite: %v", err)
	}

	if err := runner.RunTestSuite(suite); err != nil {
		t.Fatalf("Failed to run test suite: %v", err)
	}

	summary := runner.GetSummary()
	if summary.Failed > 0 {
		t.Errorf("Plugin tests failed: %d passed, %d failed", summary.Passed, summary.Failed)
		
		for _, result := range runner.GetResults() {
			if !result.Passed {
				t.Errorf("Test '%s' failed: %s", result.Name, result.Error)
			}
		}
	}
}

// DiscoverTestSuites finds all test suite files in a directory
func DiscoverTestSuites(dir string) ([]string, error) {
	var testSuites []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && (filepath.Ext(path) == ".json") {
			// Check if it's a test suite file by looking for "test" in the name
			base := filepath.Base(path)
			if filepath.Ext(base) == ".json" && 
			   (filepath.Base(filepath.Dir(path)) == "tests" || 
			    strings.Contains(base, "test")) {
				testSuites = append(testSuites, path)
			}
		}
		
		return nil
	})
	
	return testSuites, err
}