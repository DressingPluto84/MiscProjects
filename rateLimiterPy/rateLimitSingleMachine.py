import time
import math
from collections import deque
import threading

class RateLimiter:
    def __init__(self, policy, capacity, window, reset):
        if policy not in ["window", "bucket"]:
            raise ValueError("Unknown policy")
        if reset <= 0 or capacity <= 0:
            raise ValueError("Invalid token refresh rate")

        self.policy = policy
        self.state = {}
        self.window_ms = window # reqs in window ms
        self.reset_rate = reset #token / s
        self.capacity = capacity #reqs
        self.lock = threading.Lock()

    def handleRequest(self, req):
        with self.lock:
            if self.policy == "bucket":
                res = self.bucket(req)
            else:
                res = self.window(req)

        return res

    def togglePolicy(self):
        with self.lock:
            if self.policy == "bucket":
                self.policy = "window"
            else:
                self.policy = "bucket"

            self.state = {}
        return self.policy

    def window(self, req):
        if req not in self.state:
            self.state[req] = deque()

        t = time.monotonic_ns() // 1_000_000
        q = self.state[req]
        while q and t - q[0] > self.window_ms:
            q.popleft()

        if len(q) < self.capacity:
            q.append(t)
            self.state[req] = q
            return (True, self.capacity - len(q), self.window_ms - (t - q[0]))
        
        self.state[req] = q
        return (False, 0, self.window_ms - (t - q[0]))

    def bucket(self, req):
        t = time.monotonic_ns() // 1_000_000
        if req not in self.state:
            self.state[req] = [self.capacity, t]

        cap, last = self.state[req]
        cap = min(self.capacity, cap + (self.reset_rate / 1000) * (t - last))

        if cap < 1:
            self.state[req] = [cap, t]
            needed = (1 - cap) / (self.reset_rate / 1000)
            return (False, 0, math.ceil(needed))
        cap -= 1
        self.state[req] = [cap, t]
        return (True, math.floor(cap), 0)
