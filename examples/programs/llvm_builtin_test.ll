source_filename = "llvm_builtin_test.alas"

@0 = constant [30 x i8] c"Testing LLVM builtin support!\00"

declare i8* (i8*, i64) @alas_gc_alloc_array()

declare i8* (i8*, i64) @alas_gc_alloc_map()

declare void (i64) @alas_gc_retain()

declare void (i64) @alas_gc_release()

declare i8* (i8*, i64) @alas_gc_array_get()

declare i8* (i8*, i8*) @alas_gc_map_get()

declare void () @alas_gc_run()

declare void ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_io_print()

declare { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_math_sqrt()

declare { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_math_abs()

declare { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_collections_length()

declare { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_string_toUpper()

declare { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_type_typeOf()

define void () @main() {
entry:
	%0 = getelementptr [30 x i8], [30 x i8]* @0, i64 0, i64 0
	%1 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%2 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %1, i32 0, i32 0
	%3 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %1, i32 0, i32 1
	store i32 2, i32* %2
	%4 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %3, i32 0, i32 2
	store i8* %0, i8** %4
	%5 = call void ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_io_print({ i32, { i64, double, i8*, i8*, i8* } }* %1)
	%6 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%7 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %6, i32 0, i32 0
	%8 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %6, i32 0, i32 1
	store i32 0, i32* %7
	%9 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %8, i32 0, i32 0
	store i64 16, i64* %9
	%10 = call { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_math_sqrt({ i32, { i64, double, i8*, i8*, i8* } }* %6)
	%11 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%12 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %11, i32 0, i32 0
	%13 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %11, i32 0, i32 1
	store i32 1, i32* %12
	%14 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %13, i32 0, i32 1
	store double 4.0, double* %14
	%15 = call void ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_io_print({ i32, { i64, double, i8*, i8*, i8* } }* %11)
	ret void
}
