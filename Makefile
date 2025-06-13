.PHONY: all build test clean validate-example run-example build-stdlib

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

# Run tests
test:
	go test ./tests/... -v

# Clean build artifacts
clean:
	rm -rf bin/

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

# Test LLVM builtin compilation
test-llvm-builtin: build
	@echo "Testing LLVM builtin support..."
	@./bin/alas-compile -file examples/programs/llvm_builtin_test.alas.json -o examples/programs/llvm_builtin_test.ll
	@echo "Generated LLVM IR for builtin test"

# Test comprehensive LLVM builtin compilation
test-llvm-comprehensive: build
	@echo "Testing comprehensive LLVM builtin support..."
	@./bin/alas-compile -file examples/programs/comprehensive_builtin_test.alas.json -o examples/programs/comprehensive_builtin_test.ll
	@echo "Generated comprehensive LLVM IR for all builtin functions"

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