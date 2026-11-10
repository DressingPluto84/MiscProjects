#include <cassert>
#include <iostream>
#include "sharedptr.h"
#include "weakptr.h"

struct Tracker {
    static int alive;
    int value{};

    Tracker(int v=0): value(v) { ++alive; }
    ~Tracker() { --alive; }
};

int Tracker::alive = 0;

void test_basic_shared() {
    shared_ptr<int> p(new int(5));
    assert(p.use_count() == 1);
    assert(*p == 5);
}

void test_shared_copy() {
    shared_ptr<int> a(new int(3));
    shared_ptr<int> b = a;
    assert(a.use_count() == 2);
    assert(b.use_count() == 2);
}

void test_shared_move() {
    shared_ptr<int> a(new int(7));
    shared_ptr<int> b = std::move(a);
    assert(b.use_count() >= 1);
}

void test_weak_from_shared() {
    shared_ptr<int> sp(new int(10));
    weak_ptr<int> wp(sp);
    assert(!wp.expired());
}

void test_lock_success() {
    shared_ptr<int> sp(new int(20));
    weak_ptr<int> wp(sp);

    auto locked = wp.lock();
    assert(locked.use_count() >= 1);
    assert(*locked == 20);
}

void test_lock_fail() {
    weak_ptr<int> wp;

    {
        shared_ptr<int> sp(new int(30));
        wp = sp;
    }

    assert(wp.expired());
}

void test_lifetime_delete() {
    {
        shared_ptr<Tracker> sp(new Tracker(1));
        assert(Tracker::alive == 1);
    }
    assert(Tracker::alive == 0);
}

void test_weak_does_not_keep_alive() {
    weak_ptr<Tracker> wp;

    {
        shared_ptr<Tracker> sp(new Tracker(2));
        wp = sp;
        assert(!wp.expired());
    }

    assert(wp.expired());
    assert(Tracker::alive == 0);
}

void test_multiple_weak() {
    shared_ptr<int> sp(new int(100));
    weak_ptr<int> w1(sp);
    weak_ptr<int> w2(sp);

    assert(!w1.expired());
    assert(!w2.expired());
}

void test_reset() {
    shared_ptr<int> sp(new int(1));
    sp.reset(new int(2));
    assert(*sp == 2);
}

void test_swap() {
    shared_ptr<int> a(new int(1));
    shared_ptr<int> b(new int(2));

    a.swap(b);

    assert(*a == 2);
    assert(*b == 1);
}

void test_nested_locking() {
    shared_ptr<int> sp(new int(55));
    weak_ptr<int> wp(sp);

    auto l1 = wp.lock();
    auto l2 = wp.lock();

    assert(l1.use_count() >= 2);
    assert(l2.use_count() >= 2);
}

void test_copy_weak() {
    shared_ptr<int> sp(new int(5));
    weak_ptr<int> w1(sp);
    weak_ptr<int> w2 = w1;

    assert(!w2.expired());
}

void test_move_weak() {
    shared_ptr<int> sp(new int(9));
    weak_ptr<int> w1(sp);
    weak_ptr<int> w2(std::move(w1));

    assert(!w2.expired());
}

void test_owner_equality() {
    shared_ptr<int> a(new int(1));
    shared_ptr<int> b = a;

    assert(a.owner_equal(b));
}

void stress_test_lifecycle() {
    for (int i=0;i<1000;i++) {
        shared_ptr<int> sp(new int(i));
        weak_ptr<int> wp(sp);

        {
            auto l = wp.lock();
            if (!wp.expired())
                assert(*l == i);
        }
    }
}

int main() {
    std::cout << "Running tests...\n";

    test_basic_shared();
    test_shared_copy();
    test_shared_move();
    test_weak_from_shared();
    test_lock_success();
    test_lock_fail();
    test_lifetime_delete();
    test_weak_does_not_keep_alive();
    test_multiple_weak();
    test_reset();
    test_swap();
    test_nested_locking();
    test_copy_weak();
    test_move_weak();
    test_owner_equality();
    stress_test_lifecycle();

    std::cout << "All tests passed.\n";
}

