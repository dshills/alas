source_filename = "hello.alas"

@0 = constant [13 x i8] c"Hello, ALaS!\00"

define i8* () @main() {
entry:
	%0 = getelementptr [13 x i8], [13 x i8]* @0, i64 0, i64 0
	ret i8* %0
}
