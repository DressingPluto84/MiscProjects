#include <iostream>

template <typename T>
class Uptr {
	private:
		T* m_ptr{};

	public:
		Uptr(): m_ptr{nullptr} {}

		Uptr(T* val): m_ptr{val} 
			{}

		Uptr(const Uptr&) = delete;
		
		Uptr& operator=(const Uptr&) = delete;
		
		~Uptr() noexcept {
			delete m_ptr;
		}

		Uptr& operator=(Uptr&& pt) noexcept {
			if (this != &pt) {
				if (m_ptr != nullptr) {
					delete m_ptr;
				}
				m_ptr = pt.m_ptr;
				pt.m_ptr = nullptr;
			}

			return *this;
		}

		Uptr (Uptr&& pt) noexcept {
			m_ptr = pt.m_ptr;
			pt.m_ptr = nullptr;
		}

		const T& operator*() const {
			return *m_ptr;
		}
		
		const T* operator->() const {
			return m_ptr;
		}

		T& operator*() {
			return *m_ptr;
		}
		
		T* operator->() {
			return m_ptr;
		}
};

class Foo {
	public:
		int m_x{};

		Foo(int x): m_x{x} 
		{}
};

int main() {
	auto ptr = std::unique_ptr<int>(new int{5});
	auto zest = std::unique_ptr<int>(new int{2});
	zest = ptr;

	std::cout << *zest << "\n";

	return 0;
}
