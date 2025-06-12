source_filename = "fibonacci.alas"

define i64 (i64) @fibonacci(i64 %n) {
entry:
	%0 = icmp sle i64 %n, 1
	br i1 %0, label %if.then, label %if.else

if.then:
	ret i64 %n

if.else:
	%1 = sub i64 %n, 1
	%2 = call i64 (i64) @fibonacci(i64 %1)
	%3 = sub i64 %n, 2
	%4 = call i64 (i64) @fibonacci(i64 %3)
	%5 = add i64 (i64) %2, %4
	ret i64 (i64) %5

if.end:
	unreachable
}

define i64 () @main() {
entry:
	%0 = call i64 (i64) @fibonacci(i64 10)
	ret i64 (i64) %0
}
