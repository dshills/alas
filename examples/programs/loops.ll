source_filename = "loops.alas"

define i64 (i64) @sum_to_n(i64 %n) {
entry:
	br label %while.cond

while.cond:
	%0 = icmp sle i64 1, %n
	br i1 %0, label %while.body, label %while.end

while.body:
	%1 = add i64 0, 1
	%2 = add i64 1, 1
	br label %while.cond

while.end:
	ret i64 %1
}

define i64 () @main() {
entry:
	%0 = call i64 (i64) @sum_to_n(i64 10)
	ret i64 (i64) %0
}
