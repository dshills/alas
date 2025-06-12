source_filename = "factorial.alas"

define i64 (i64) @factorial(i64 %n) {
entry:
	%0 = icmp sle i64 %n, 1
	br i1 %0, label %if.then, label %if.else

if.then:
	ret i64 1

if.else:
	%1 = sub i64 %n, 1
	%2 = call i64 (i64) @factorial(i64 %1)
	%3 = mul i64 %n, %2
	ret i64 %3

if.end:
	unreachable
}

define i64 () @main() {
entry:
	%0 = call i64 (i64) @factorial(i64 5)
	ret i64 (i64) %0
}
