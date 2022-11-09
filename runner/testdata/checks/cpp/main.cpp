#include <iostream>

int main(void) {
	int arr[10] = {};

	int n;
	std::cin >> n;

	// [CXX-S1001]: "Array-index variable with possibly unchecked bounds"
	std::cout << arr[n] << std::endl;

	arr[0] = "1"; // [CXX-S1002]: "Unchecked parameter value used in array index"

	return 0;
}
