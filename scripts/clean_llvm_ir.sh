#!/bin/bash

# Clean LLVM IR script
# This script normalizes LLVM IR output from alas-compile for compatibility
# with standard LLVM toolchain

set -e
set -o pipefail

input_file="$1"
output_file="$2"

if [ -z "$input_file" ] || [ -z "$output_file" ]; then
    echo "Usage: $0 <input.ll> <output.ll>"
    exit 1
fi

# Use LLVM's opt tool for proper IR normalization if available
if command -v opt &> /dev/null; then
    # Use LLVM's optimizer with -O0 to just clean/normalize the IR
    opt -S -O0 "$input_file" -o "$output_file"
else
    # Fallback to sed-based cleaning with more robust patterns
    # These transformations handle specific ALaS codegen quirks
    cat "$input_file" | \
    # Fix function declaration syntax
    sed -E 's/declare ([^(]*) \(([^)]*)\) (@[^(]*)\(\)/declare \1 \3(\2)/g' | \
    # Fix main function definition
    sed -E 's/define void \(\) @main\(\)/define void @main()/' | \
    # Remove unnecessary assignment from void calls
    sed -E 's/%[0-9]+ = call void.*@alas_builtin_io_print/call void @alas_builtin_io_print/g' | \
    # Fix bitcast syntax issues
    sed -E 's/bitcast i8\* \(i8\*\)/bitcast i8*/g' | \
    sed -E 's/call i8\* \(i8\*\)/call i8*/g' \
    > "$output_file"
fi

echo "Cleaned LLVM IR: $output_file"