import redis
import time
import math
from collections import deque
import threading
import json

class RateLimiter:
    def __init__(self, policy, capacity, window, reset, redis):
        if policy not in ["window", "bucket"]:
            raise ValueError("Unknown policy")
        if reset <= 0 or capacity <= 0:
            raise ValueError("Invalid token refresh rate")

        self.policy = policy
        self.window_ms = window # reqs in window ms
        self.reset_rate = reset #token / s
        self.capacity = capacity #reqs
        self.r = redis
        self.lock = self.r.lock("lock", timeout=5)

    def handleRequest(self, req):
        with self.lock:
            if self.policy == "bucket":
                req = ",".join(req + ("bucket",))
                res = self.bucket(req)
            else:
                req = ",".join(req + ("window",))
                res = self.window(req)

        return res

    def togglePolicy(self):
        with self.lock:
            if self.policy == "bucket":
                self.policy = "window"
            else:
                self.policy = "bucket"

        return self.policy

    def window(self, req):
        t = time.monotonic_ns() // 1_000_000
        while self.r.exists(req) and t - int(self.r.lindex(req, 0)) > self.window_ms:
            self.r.lpop(req)

        if self.r.llen(req) < self.capacity:
            self.r.rpush(req, t)
            return (True, self.capacity - self.r.llen(req), self.window_ms - (t - int(self.r.lindex(req, 0))))
        
        return (False, 0, self.window_ms - (t - int(self.r.lindex(req, 0))))

    def bucket(self, req):
        t = time.monotonic_ns() // 1_000_000
        if not self.r.get(req):
            self.r.set(req, f"{self.capacity},{t}", ex=60)

        val = self.r.get(req).split(",")
        cap, last = float(val[0]), int(val[1])
        cap = min(self.capacity, cap + (self.reset_rate / 1000) * (t - last))

        if cap < 1:
            self.r.set(req, f"{cap},{t}")
            needed = (1 - cap) / (self.reset_rate / 1000)
            return (False, 0, math.ceil(needed))
        cap -= 1
        self.r.set(req, f"{cap},{t}")
        return (True, math.floor(cap), 0)

r = redis.Redis(host='localhost', port=6379, decode_responses=True)
rl = RateLimiter("window", 4, 10000, 0.5, r)
for i in range(6):
    print(rl.handleRequest(("u1", "/home")))
