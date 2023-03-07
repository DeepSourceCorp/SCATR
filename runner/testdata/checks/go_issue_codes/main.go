// scatr-ignore: SCC-U1000
package main

// [GO-C5001]: 9 "Redundant type in variable declaration"
var foo int = 10

func bar() {
	a := 10
	return
	// [VET-V0002]: "Useless assignment"; [SCC-U1000]: "Code is unused"
	a = a
}
