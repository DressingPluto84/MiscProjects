#pragma once

#include <iostream>

template <typename T>
class weak_ptr;

template <typename T>
class shared_ptr {
	private:
		T* m_ptr;
		int64_t* m_counter{};
		int64_t* m_wcounter{};
	
		shared_ptr(T* ptr, int64_t* c, int64_t* wc): m_ptr{ptr}, m_counter{c}, m_wcounter{wc} {}
			
	public:
		shared_ptr(): m_ptr{nullptr}, m_counter{new int64_t(0)}, m_wcounter{new int64_t(0)} {}

		shared_ptr(T* ptr): m_ptr{ptr}, m_counter{new int64_t(1)}, m_wcounter{new int64_t{1}} {}

		~shared_ptr() {
			*m_counter -= 1;

			if (*m_counter == 0) {
				delete m_ptr;
			}
			if (*m_wcounter == 0) {
				delete m_counter;
				delete m_wcounter;
			}
		}

		shared_ptr(const shared_ptr& p) {
			m_ptr = p.m_ptr;
			m_counter = p.m_counter;
			m_wcounter = p.m_wcounter;
			*m_counter += 1;
		}

		shared_ptr& operator=(const shared_ptr& p) {
			if (m_ptr == p.m_ptr) {
				return *this;
			}

			*m_counter -= 1;
			if (*m_counter == 0) {
				delete m_ptr;
			}
			if (*m_wcounter == 0) {
				delete m_counter;
				delete m_wcounter;
			}
			m_ptr = p.m_ptr;
			m_counter = p.m_counter;
			m_wcounter = p.m_wcounter;
			*m_counter += 1;

			return *this;
		}

		shared_ptr(shared_ptr&& p) noexcept {
			m_ptr = p.m_ptr;
			m_counter = p.m_counter;
			m_wcounter = p.m_wcounter;
			*m_counter += 1;
			p.m_ptr = nullptr;
		}

		shared_ptr& operator=(shared_ptr&& p) noexcept {
			if (m_ptr == p.m_ptr) {
				return this;
			}
		
			*m_counter -= 1;
			if (*m_counter == 0) {
				delete m_ptr;
			}
			if (*m_wcounter == 0) {
				delete m_counter;
				delete m_wcounter;
			}
			m_ptr = p.m_ptr;
			m_counter = p.m_counter;
			m_wcounter = p.m_wcounter;
			*m_counter += 1;
			p.m_ptr = nullptr;
		}

		void reset(T* newp=nullptr) {
			*m_counter -= 1;
			if (*m_counter == 0) {
				delete m_ptr;
			}
			if (m_wcounter == 0) {
				delete m_counter;
				delete m_wcounter;
			}

			m_counter = new int64_t{0};
			if (newp != nullptr) {
				*m_counter += 1;
			}
			m_ptr = newp;
		}

		void swap(shared_ptr& p) {
			if (this == &p) {
				return;
			}

			T* val{m_ptr};
			m_ptr = p.m_ptr;
			p.m_ptr = val;

			int64_t c{*m_counter};
			*m_counter = *p.m_counter;
			*m_wcounter = *p.m_wcounter;
			*p.m_counter = c;
		}

		uint64_t use_count() {
			return *m_counter;
		}

		T* get() {
			return m_ptr;
		}

		explicit operator bool() const noexcept {
			return m_ptr != nullptr;
		}

		bool owner_equal(const shared_ptr& other) const noexcept {
			return m_ptr == other.m_ptr;
		}

		T& operator*() {
			return *m_ptr;
		}

		T* operator->() {
			return m_ptr;
		}

		const T& operator*() const {
			return *m_ptr;
		}

		const T* operator->() const {
			return m_ptr;
		}

		template <typename U>
		friend void print(const shared_ptr<U>& p);

		friend class weak_ptr<T>;
};

template <typename T>
void print(const shared_ptr<T>& p) {
	if (p.m_ptr != nullptr) {
		std::cout << *p.m_ptr << "\n";
	}
	std::cout << p.m_ptr  << " " << *(p.m_counter) << " \n";
}

