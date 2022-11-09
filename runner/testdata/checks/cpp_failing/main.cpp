#include <iostream>

int main(void) {
	int arr[10] = {};

	int n;
	std::cin >> n;

	std::cout << arr[n] << std::endl;

	arr[0] = "1";

	return 0;
}
