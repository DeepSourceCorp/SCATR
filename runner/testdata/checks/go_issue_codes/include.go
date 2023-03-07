// scatr-check: GO-W1008, GO-W1000
package main

import (
	"fmt"
)

func _() {
	// [GO-W1008]; [GO-W1009]
	fmt.Println("Hello World")
	// [GO-W1000]
	fmt.Println("Foo")
}
