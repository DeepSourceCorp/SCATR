package code

// [GO-C5001]: 9 "Redundant type in variable declaration"
var foo_ int = 10

func bar_() {
	a := 10
	// [VET-V0002]: "Useless assignment"
	a = a

	// [VET-V0002]
	a = a

	// [VET-V0002]
	a = a
}
