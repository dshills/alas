source_filename = "comprehensive_builtin_test.alas"

@0 = constant [39 x i8] c"Testing comprehensive builtin support!\00"
@1 = constant [6 x i8] c"hello\00"
@2 = constant [5 x i8] c"test\00"
@3 = constant [10 x i8] c"RESULT\5Cx00"
@4 = constant [9 x i8] c"float\5Cx00"

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
	%0 = getelementptr [39 x i8], [39 x i8]* @0, i64 0, i64 0
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
	store i32 0, i32* %12
	%14 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %13, i32 0, i32 0
	store i64 -5, i64* %14
	%15 = call { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_math_abs({ i32, { i64, double, i8*, i8*, i8* } }* %11)
	%16 = getelementptr [6 x i8], [6 x i8]* @1, i64 0, i64 0
	%17 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%18 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %17, i32 0, i32 0
	%19 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %17, i32 0, i32 1
	store i32 2, i32* %18
	%20 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %19, i32 0, i32 2
	store i8* %16, i8** %20
	%21 = call { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_collections_length({ i32, { i64, double, i8*, i8*, i8* } }* %17)
	%22 = getelementptr [5 x i8], [5 x i8]* @2, i64 0, i64 0
	%23 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%24 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %23, i32 0, i32 0
	%25 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %23, i32 0, i32 1
	store i32 2, i32* %24
	%26 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %25, i32 0, i32 2
	store i8* %22, i8** %26
	%27 = call { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_string_toUpper({ i32, { i64, double, i8*, i8*, i8* } }* %23)
	%28 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%29 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %28, i32 0, i32 0
	%30 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %28, i32 0, i32 1
	store i32 0, i32* %29
	%31 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %30, i32 0, i32 0
	store i64 42, i64* %31
	%32 = call { i32, { i64, double, i8*, i8*, i8* } } ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_type_typeOf({ i32, { i64, double, i8*, i8*, i8* } }* %28)
	%33 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%34 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %33, i32 0, i32 0
	%35 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %33, i32 0, i32 1
	store i32 1, i32* %34
	%36 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %35, i32 0, i32 1
	store double 4.0, double* %36
	%37 = call void ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_io_print({ i32, { i64, double, i8*, i8*, i8* } }* %33)
	%38 = alloca { i32, { i64, double, i8*, i8*, i8* } }
	%39 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %38, i32 0, i32 0
	%40 = getelementptr { i32, { i64, double, i8*, i8*, i8* } }, { i32, { i64, double, i8*, i8*, i8* } }* %38, i32 0, i32 1
	store i32 1, i32* %39
	%41 = getelementptr { i64, double, i8*, i8*, i8* }, { i64, double, i8*, i8*, i8* }* %40, i32 0, i32 1
	store double 1.0, double* %41
	%42 = call void ({ i32, { i64, double, i8*, i8*, i8* } }*) @alas_builtin_io_print({ i32, { i64, double, i8*, i8*, i8* } }* %38)
	ret void
}
