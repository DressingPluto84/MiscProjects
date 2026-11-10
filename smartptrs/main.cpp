#include "sharedptr.h"
#include "weakptr.h"

shared_ptr<int> foo() {
	shared_ptr<int> p = shared_ptr<int>(new int(4));
	return p;
} 

int main() {
	auto ptr = std::weak_ptr<int>();
	auto pt2 = std::make_shared<int>(10);
	ptr = pt2;
	auto pt3 = ptr;

	std::cout << pt3.use_count() << " " << *pt2 << "\n";

	return 0;
}
