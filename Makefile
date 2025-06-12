.PHONY: all build test clean validate-example run-example

# Build all binaries
all: build

build:
	go build -o bin/alas-validate ./cmd/alas-validate
	go build -o bin/alas-run ./cmd/alas-run
	go build -o bin/alas-compile ./cmd/alas-compile

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

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...