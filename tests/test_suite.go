package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestSuite manages comprehensive testing with cleanup and reporting
type TestSuite struct {
	tempDirs    []string
	tempFiles   []string
	startTime   time.Time
	cleanupDone bool
}

// NewTestSuite creates a new test suite instance
func NewTestSuite() *TestSuite {
	return &TestSuite{
		tempDirs:  make([]string, 0),
		tempFiles: make([]string, 0),
		startTime: time.Now(),
	}
}

// Setup initializes the test suite
func (ts *TestSuite) Setup(t *testing.T) {
	t.Helper()
	ts.startTime = time.Now()
	
	// Register cleanup to be called at the end
	t.Cleanup(func() {
		ts.Cleanup(t)
	})
	
	t.Logf("Test suite setup completed at %v", ts.startTime)
}

// RegisterTempDir registers a temporary directory for cleanup
func (ts *TestSuite) RegisterTempDir(dir string) {
	ts.tempDirs = append(ts.tempDirs, dir)
}

// RegisterTempFile registers a temporary file for cleanup
func (ts *TestSuite) RegisterTempFile(file string) {
	ts.tempFiles = append(ts.tempFiles, file)
}

// CreateTempDir creates and registers a temporary directory
func (ts *TestSuite) CreateTempDir(t *testing.T, pattern string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	ts.RegisterTempDir(dir)
	return dir
}

// CreateTempFile creates and registers a temporary file
func (ts *TestSuite) CreateTempFile(t *testing.T, dir, pattern string) *os.File {
	t.Helper()
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	ts.RegisterTempFile(file.Name())
	return file
}

// Cleanup removes all temporary files and directories
func (ts *TestSuite) Cleanup(t *testing.T) {
	if ts.cleanupDone {
		return
	}
	ts.cleanupDone = true
	
	t.Helper()
	
	var errors []string
	
	// Clean up temporary files
	for _, file := range ts.tempFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("failed to remove temp file %s: %v", file, err))
		}
	}
	
	// Clean up temporary directories
	for _, dir := range ts.tempDirs {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("failed to remove temp dir %s: %v", dir, err))
		}
	}
	
	// Clean up any leftover .ll and .o files in current directory
	if err := ts.cleanupCompilerArtifacts(t); err != nil {
		errors = append(errors, fmt.Sprintf("failed to cleanup compiler artifacts: %v", err))
	}
	
	duration := time.Since(ts.startTime)
	t.Logf("Test suite cleanup completed in %v", duration)
	
	if len(errors) > 0 {
		t.Logf("Cleanup warnings: %s", strings.Join(errors, "; "))
	}
}

// cleanupCompilerArtifacts removes LLVM and object files from the current directory
func (ts *TestSuite) cleanupCompilerArtifacts(t *testing.T) error {
	t.Helper()
	
	// Find and remove .ll files
	llFiles, err := filepath.Glob("*.ll")
	if err != nil {
		return fmt.Errorf("failed to glob .ll files: %v", err)
	}
	
	for _, file := range llFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove %s: %v", file, err)
		}
	}
	
	// Find and remove .o files
	objFiles, err := filepath.Glob("*.o")
	if err != nil {
		return fmt.Errorf("failed to glob .o files: %v", err)
	}
	
	for _, file := range objFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove %s: %v", file, err)
		}
	}
	
	// Find and remove executable files (common names)
	executablePatterns := []string{"test", "main", "factorial", "fibonacci", "hello"}
	for _, pattern := range executablePatterns {
		if info, err := os.Stat(pattern); err == nil && !info.IsDir() {
			// Check if it's likely an executable (no extension on Unix-like systems)
			if runtime.GOOS != "windows" || strings.HasSuffix(pattern, ".exe") {
				if err := os.Remove(pattern); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: failed to remove executable %s: %v", pattern, err)
				}
			}
		}
	}
	
	return nil
}

// CheckMemoryUsage reports current memory usage (for monitoring memory leaks)
func (ts *TestSuite) CheckMemoryUsage(t *testing.T, label string) {
	t.Helper()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	t.Logf("Memory usage at %s: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d",
		label,
		m.Alloc/1024,
		m.TotalAlloc/1024,
		m.Sys/1024,
		m.NumGC,
	)
}

// ForceGC forces garbage collection and waits for it to complete
func (ts *TestSuite) ForceGC(t *testing.T) {
	t.Helper()
	
	before := time.Now()
	runtime.GC()
	runtime.GC() // Run twice to ensure finalization
	duration := time.Since(before)
	
	t.Logf("Forced GC completed in %v", duration)
}

// ValidateTestEnvironment checks that the test environment is properly set up
func (ts *TestSuite) ValidateTestEnvironment(t *testing.T) {
	t.Helper()
	
	// Check current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	
	// If we're in the tests directory, change to parent directory
	if strings.HasSuffix(cwd, "/tests") || strings.HasSuffix(cwd, "\\tests") {
		parentDir := filepath.Dir(cwd)
		if err := os.Chdir(parentDir); err != nil {
			t.Fatalf("Failed to change to parent directory: %v", err)
		}
		cwd = parentDir
		t.Logf("Changed working directory to: %s", cwd)
	}
	
	// Verify we're in the right project directory
	if !strings.HasSuffix(cwd, "alas") && !strings.Contains(cwd, "alas") {
		t.Logf("Warning: May not be in ALaS project directory. CWD: %s", cwd)
	}
	
	// Check for required directories
	requiredDirs := []string{"internal", "examples", "tests"}
	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Logf("Warning: Required directory %s not found", dir)
		}
	}
	
	// Check for required example files
	requiredExamples := []string{
		"examples/programs/hello.alas.json",
		"examples/programs/factorial.alas.json",
		"examples/programs/fibonacci.alas.json",
	}
	
	for _, example := range requiredExamples {
		if _, err := os.Stat(example); os.IsNotExist(err) {
			t.Logf("Warning: Required example file %s not found", example)
		}
	}
	
	t.Logf("Test environment validation completed. CWD: %s", cwd)
}

// BenchmarkHelper provides utilities for benchmarking tests
type BenchmarkHelper struct {
	suite *TestSuite
}

// NewBenchmarkHelper creates a new benchmark helper
func (ts *TestSuite) NewBenchmarkHelper() *BenchmarkHelper {
	return &BenchmarkHelper{suite: ts}
}

// RunBenchmark runs a benchmark with cleanup
func (bh *BenchmarkHelper) RunBenchmark(b *testing.B, name string, fn func(b *testing.B)) {
	b.Helper()
	
	b.Run(name, func(b *testing.B) {
		// Setup cleanup for benchmark
		b.Cleanup(func() {
			bh.cleanupForBenchmark(b)
		})
		
		// Run the actual benchmark
		fn(b)
	})
}

// cleanupForBenchmark performs cleanup operations for benchmarks
func (bh *BenchmarkHelper) cleanupForBenchmark(b *testing.B) {
	b.Helper()
	
	suite := bh.suite
	if suite.cleanupDone {
		return
	}
	suite.cleanupDone = true
	
	var errors []string
	
	// Clean up temporary files
	for _, file := range suite.tempFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("failed to remove temp file %s: %v", file, err))
		}
	}
	
	// Clean up temporary directories
	for _, dir := range suite.tempDirs {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("failed to remove temp dir %s: %v", dir, err))
		}
	}
	
	// Clean up compiler artifacts
	llFiles, _ := filepath.Glob("*.ll")
	for _, file := range llFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			b.Logf("Warning: failed to remove %s: %v", file, err)
		}
	}
	
	objFiles, _ := filepath.Glob("*.o")
	for _, file := range objFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			b.Logf("Warning: failed to remove %s: %v", file, err)
		}
	}
	
	executables, _ := filepath.Glob("test_*")
	for _, file := range executables {
		if info, err := os.Stat(file); err == nil && info.Mode().IsRegular() && info.Mode()&0111 != 0 {
			if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
				b.Logf("Warning: failed to remove %s: %v", file, err)
			}
		}
	}
	
	duration := time.Since(suite.startTime)
	b.Logf("Benchmark cleanup completed in %v", duration)
	
	if len(errors) > 0 {
		b.Logf("Cleanup warnings: %s", strings.Join(errors, "; "))
	}
	
	// Reset the suite for potential reuse
	suite.tempFiles = suite.tempFiles[:0]
	suite.tempDirs = suite.tempDirs[:0]
	suite.cleanupDone = false
}

// TestResult represents the result of a test execution
type TestResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Error    error
	Message  string
}

// TestResults collects and reports test results
type TestResults struct {
	Results []TestResult
	Summary map[string]int
}

// NewTestResults creates a new test results collector
func NewTestResults() *TestResults {
	return &TestResults{
		Results: make([]TestResult, 0),
		Summary: make(map[string]int),
	}
}

// AddResult adds a test result
func (tr *TestResults) AddResult(result TestResult) {
	tr.Results = append(tr.Results, result)
	
	if result.Passed {
		tr.Summary["passed"]++
	} else {
		tr.Summary["failed"]++
	}
}

// GenerateReport generates a comprehensive test report
func (tr *TestResults) GenerateReport(t *testing.T) {
	t.Helper()
	
	total := len(tr.Results)
	passed := tr.Summary["passed"]
	failed := tr.Summary["failed"]
	
	t.Logf("=== TEST SUMMARY ===")
	t.Logf("Total tests: %d", total)
	t.Logf("Passed: %d", passed)
	t.Logf("Failed: %d", failed)
	
	if failed > 0 {
		t.Logf("=== FAILED TESTS ===")
		for _, result := range tr.Results {
			if !result.Passed {
				t.Logf("FAIL: %s - %s", result.Name, result.Message)
				if result.Error != nil {
					t.Logf("  Error: %v", result.Error)
				}
			}
		}
	}
	
	// Calculate average duration
	if total > 0 {
		var totalDuration time.Duration
		for _, result := range tr.Results {
			totalDuration += result.Duration
		}
		avgDuration := totalDuration / time.Duration(total)
		t.Logf("Average test duration: %v", avgDuration)
	}
	
	t.Logf("=== END SUMMARY ===")
}

// ParallelTestRunner runs tests in parallel with proper cleanup
type ParallelTestRunner struct {
	suite   *TestSuite
	maxGoroutines int
}

// NewParallelTestRunner creates a new parallel test runner
func (ts *TestSuite) NewParallelTestRunner(maxGoroutines int) *ParallelTestRunner {
	if maxGoroutines <= 0 {
		maxGoroutines = runtime.NumCPU()
	}
	
	return &ParallelTestRunner{
		suite:         ts,
		maxGoroutines: maxGoroutines,
	}
}

// RunTests runs multiple tests in parallel
func (ptr *ParallelTestRunner) RunTests(t *testing.T, tests []func(*testing.T)) {
	t.Helper()
	
	// Create a semaphore to limit concurrent goroutines
	semaphore := make(chan struct{}, ptr.maxGoroutines)
	results := make(chan TestResult, len(tests))
	
	// Start all tests
	for i, testFunc := range tests {
		go func(index int, fn func(*testing.T)) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore
			
			start := time.Now()
			
			// Create a sub-test
			testName := fmt.Sprintf("ParallelTest_%d", index)
			result := TestResult{
				Name:     testName,
				Duration: time.Since(start),
			}
			
			// Run the test in a sub-test to capture results
			t.Run(testName, func(subT *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						result.Passed = false
						result.Error = fmt.Errorf("panic: %v", r)
						result.Message = "Test panicked"
					}
					result.Duration = time.Since(start)
					results <- result
				}()
				
				fn(subT)
				result.Passed = !subT.Failed()
			})
		}(i, testFunc)
	}
	
	// Wait for all tests to complete
	testResults := NewTestResults()
	for i := 0; i < len(tests); i++ {
		result := <-results
		testResults.AddResult(result)
	}
	
	// Generate report
	testResults.GenerateReport(t)
}