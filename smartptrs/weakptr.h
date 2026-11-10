#pragma once

#include <iostream>
#include "sharedptr.h"

template <typename T>
class weak_ptr {
	private:
		T* m_ptr{};
		int64_t* m_counter{};
		int64_t* m_wcounter{};

	public:
		weak_ptr(): m_ptr{nullptr}, m_counter{new int64_t(0)}, m_wcounter{new int64_t(0)} {} 

		weak_ptr(const shared_ptr<T>& sp): m_ptr{sp.m_ptr}, m_counter{sp.m_counter}, m_wcounter{sp.m_wcounter} {}

		~weak_ptr() {
			*m_wcounter -= 1;
			if ((m_counter == nullptr or *m_counter == 0) and *m_wcounter == 0) {
				delete m_counter;
				delete m_wcounter;
			}
		}

		weak_ptr& operator=(const shared_ptr<T>& sp) {
			if (m_counter != nullptr) {
				*m_wcounter -= 1;
				if (*m_counter == 0 and *m_wcounter == 0) {
					delete m_counter;
					delete m_wcounter;
				}
			}

			m_ptr = sp.m_ptr;
			m_counter = sp.m_counter;
			m_wcounter = sp.m_wcounter;
			return *this;
		}

		weak_ptr(const weak_ptr<T>& wp): m_ptr{wp.m_ptr}, m_counter{wp.m_counter}, m_wcounter{wp.m_wcounter} {}

		weak_ptr& operator=(const weak_ptr<T>& wp) noexcept {
			if (this == &wp) {
				return *this;
			}
			if (m_counter != nullptr) {
				*m_wcounter -= 1;
				if (*m_counter == 0 and *m_wcounter == 0) {
					delete m_counter;
					delete m_wcounter;
				}
			}

			m_ptr = wp.m_ptr;
			m_counter = wp.m_counter;
			m_wcounter = wp.m_wcounter;
			return *this;
		}

		weak_ptr(weak_ptr<T>&& wp) noexcept
			: m_ptr{wp.m_ptr},
			m_counter{wp.m_counter},
			m_wcounter{wp.m_wcounter}
		{
			wp.m_ptr = nullptr;
			wp.m_counter = nullptr;
		}

		weak_ptr& operator=(weak_ptr<T>&& wp) noexcept {
			if (m_ptr == wp.m_ptr) {
				return *this;
			}

			if (m_counter != nullptr) {
				*m_wcounter -= 1;
				if (*m_counter == 0 and *m_wcounter == 0) {
					delete m_counter;
					delete m_wcounter;
				}
			}

			m_ptr = wp.m_ptr;
			m_counter = wp.m_counter;
			m_wcounter = wp.m_wcounter;
			*m_wcounter += 1;

			wp.m_ptr = nullptr;
			wp.m_counter = nullptr;
			return *this;
		}

		weak_ptr(const shared_ptr<T>&& sp) noexcept {
			m_ptr = sp.m_ptr;
			m_counter = sp.m_counter;
			m_wcounter = sp.m_wcounter;
			*m_wcounter += 1;
		}

		weak_ptr& operator=(const shared_ptr<T>&& sp) noexcept {
			if (m_counter != nullptr) {
				*m_wcounter -= 1;
				if (*m_counter == 0 and *m_wcounter == 0) {
					delete m_counter;
					delete m_wcounter;
				}
			}

			m_ptr = sp.m_ptr;
			m_counter = sp.m_counter;
			m_wcounter = sp.m_wcounter;
			*m_wcounter += 1;

			return *this;
		}

		bool expired() {
			if (!m_counter) return true;
			return *m_counter == 0;
		}

		shared_ptr<T> lock() {
			if (!expired()) {
				*m_counter += 1;
				return shared_ptr<T>(m_ptr, m_counter, m_wcounter);
			}

			return shared_ptr<T>();
		}

		void print() {
			std::cout << m_ptr << " " << *m_counter << " " << m_wcounter << "\n";
		}
};

