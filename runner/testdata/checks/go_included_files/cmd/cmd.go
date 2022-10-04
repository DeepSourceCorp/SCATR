package cmd

// [SCC-U1000]: "func foo() is unused"
func foo() {
	a := 10
	// [VET-V0002]: "Useless assignment"
	a = a
}
