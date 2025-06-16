source_filename = "custom_types_demo.alas"

@0 = constant [4 x i8] c"Bob\00"
@1 = constant [15 x i8] c"Person created\00"

declare i8* @alas_gc_alloc_array(i8* %0, i64 %1)

declare i8* @alas_gc_alloc_map(i8* %0, i64 %1)

declare void @alas_gc_retain(i64 %0)

declare void @alas_gc_release(i64 %0)

declare i8* @alas_gc_array_get(i8* %0, i64 %1)

declare i8* @alas_gc_map_get(i8* %0, i8* %1)

declare void @alas_gc_run()

declare void @alas_builtin_io_print(i8* %0)

declare i8* @alas_builtin_math_sqrt(i8* %0)

declare i8* @alas_builtin_math_abs(i8* %0)

declare i8* @alas_builtin_math_max(i8* %0, i8* %1)

declare i8* @alas_builtin_math_min(i8* %0, i8* %1)

declare i8* @alas_builtin_collections_length(i8* %0)

declare i8* @alas_builtin_collections_contains(i8* %0, i8* %1)

declare i8* @alas_builtin_string_toUpper(i8* %0)

declare i8* @alas_builtin_string_toLower(i8* %0)

declare i8* @alas_builtin_string_length(i8* %0)

declare i8* @alas_builtin_type_typeOf(i8* %0)

declare i8* @alas_builtin_type_isInt(i8* %0)

define { i8*, i64 } @create_person(i8* %name, i64 %age) {
entry:
	%name_ptr = alloca i8*
	store i8* %name, i8** %name_ptr
	%age_ptr = alloca i64
	store i64 %age, i64* %age_ptr
	%Person_struct = alloca { i8*, i64 }
	%0 = load { i8*, i64 }, { i8*, i64 }* %Person_struct
	ret { i8*, i64 } %0
}

define i64 @main() {
entry:
	%0 = getelementptr [4 x i8], [4 x i8]* @0, i64 0, i64 0
	%1 = call { i8*, i64 } @create_person(i8* %0, i64 30)
	%2 = getelementptr [15 x i8], [15 x i8]* @1, i64 0, i64 0
	call void @alas_builtin_io_print(i8* %2)
	ret i64 0
}
