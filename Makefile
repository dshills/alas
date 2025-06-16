.PHONY: all build test test-verbose clean validate-example run-example build-stdlib compile-to-native run-compiled compare-output

# Build all binaries
all: build

build:
	go build -o bin/alas-validate ./cmd/alas-validate
	go build -o bin/alas-run ./cmd/alas-run
	go build -o bin/alas-compile ./cmd/alas-compile
	go build -o bin/alas-plugin ./cmd/alas-plugin
	go build -o bin/alas-compile-multi ./cmd/alas-compile-multi

# Build the standard library as a shared library
build-stdlib:
	go build -buildmode=c-shared -o lib/libalas_stdlib.so ./cmd/alas-stdlib
	@echo "Built shared library: lib/libalas_stdlib.so"

# Run tests (showing only failures)
test:
	@echo "Running tests..."
	@if go test ./tests/... > /tmp/test_output.txt 2>&1; then \
		echo "✓ All tests passed"; \
	else \
		echo "✗ Tests failed:"; \
		grep -E "^--- FAIL:|^\s+.*_test\.go:|^FAIL" /tmp/test_output.txt || cat /tmp/test_output.txt; \
		exit 1; \
	fi

# Run tests with verbose output
test-verbose:
	go test ./tests/... -v

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f examples/programs/*.ll
	rm -f examples/programs/*_exe
	rm -f *.ll

# Validate an example program
validate-example:
	./bin/alas-validate -file examples/programs/hello.alas.json

# Run an example program
run-example:
	./bin/alas-run -file examples/programs/hello.alas.json

# Run all examples
run-all-examples: build
	@echo "Running hello.alas.json..."
	@./bin/alas-run -file examples/programs/hello.alas.json
	@echo "\nRunning fibonacci.alas.json..."
	@./bin/alas-run -file examples/programs/fibonacci.alas.json
	@echo "\nRunning factorial.alas.json..."
	@./bin/alas-run -file examples/programs/factorial.alas.json
	@echo "\nRunning loops.alas.json..."
	@./bin/alas-run -file examples/programs/loops.alas.json

# Compile examples to LLVM IR
compile-examples: build
	@echo "Compiling examples to LLVM IR..."
	@./bin/alas-compile -file examples/programs/hello.alas.json -o examples/programs/hello.ll
	@./bin/alas-compile -file examples/programs/fibonacci.alas.json -o examples/programs/fibonacci.ll
	@./bin/alas-compile -file examples/programs/factorial.alas.json -o examples/programs/factorial.ll
	@./bin/alas-compile -file examples/programs/loops.alas.json -o examples/programs/loops.ll
	@echo "LLVM IR files generated in examples/programs/"

compile-examples-clean:
	@rm -f examples/programs/*.ll

# Test LLVM builtin compilation
test-llvm-builtin: build
	@echo "Testing LLVM builtin support..."
	@./bin/alas-compile -file examples/programs/llvm_builtin_test.alas.json -o examples/programs/llvm_builtin_test.ll
	@echo "Generated LLVM IR for builtin test"
	@rm -f examples/programs/llvm_builtin_test.ll

# Test comprehensive LLVM builtin compilation
test-llvm-comprehensive: build
	@echo "Testing comprehensive LLVM builtin support..."
	@./bin/alas-compile -file examples/programs/comprehensive_builtin_test.alas.json -o examples/programs/comprehensive_builtin_test.ll
	@echo "Generated comprehensive LLVM IR for all builtin functions"
	@rm -f examples/programs/comprehensive_builtin_test.ll

# Compile ALaS to native executable via LLVM
compile-to-native: build build-stdlib
	@echo "Compiling ALaS to native executable..."
	@./bin/alas-compile -file examples/programs/simple_builtin_test.alas.json -o examples/programs/simple_builtin_test.ll
	@echo "Generated LLVM IR"
	@clang examples/programs/simple_builtin_test.ll -L. -lalas_runtime -o examples/programs/simple_builtin_test_exe
	@echo "Linked native executable"

# Run compiled executable
run-compiled: compile-to-native
	@echo "Running compiled executable:"
	@cd examples/programs && DYLD_LIBRARY_PATH=../.. ./simple_builtin_test_exe
	@rm -f examples/programs/simple_builtin_test.ll
	@rm -f examples/programs/simple_builtin_test_exe

# Compare interpreter vs compiled output
compare-output: build run-compiled
	@echo ""
	@echo "Running with interpreter:"
	@./bin/alas-run -file examples/programs/simple_builtin_test.alas.json

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Plugin management
plugin-list: build
	./bin/alas-plugin list -path examples/plugins

plugin-create: build
	./bin/alas-plugin create test-plugin

validate-plugins: build
	@echo "Validating example plugins..."
	@./bin/alas-plugin validate examples/plugins/hello-world/plugin.json
	@./bin/alas-plugin validate examples/plugins/math-utils/plugin.json