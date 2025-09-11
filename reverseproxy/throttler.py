from random import randint
import time

class DownloadSpeedLimiter:
    def __init__(self, max_speed_bytes_per_sec):
        self.max_speed = max_speed_bytes_per_sec
        # self._start_time = time.time()
        self._prev_time = time.time() - 10
        self._estimated_speed = 0
        self._ratio = 0.5
        self._delay = 0.0

    def consumed(self, bytes_count: int):
        now = time.time()
        elapsed = now - self._prev_time
        self._prev_time = now
        self._estimated_speed = self._estimated_speed * self._ratio + (1.0 - self._ratio) * (bytes_count + 0.0)/elapsed

    def delay(self) -> float:
        newDelay = (self._estimated_speed - self.max_speed) / self.max_speed
        self._delay = self._delay + 0.2 * newDelay
        return max(0.0, self._delay)

    
if __name__ == "__main__":
    start = time.time()
    limiter = DownloadSpeedLimiter(2000)
    counterStart = time.time()
    consumed = 0
    def consume(number):
        global consumed, counterStart
        if time.time() - counterStart > 10:
            consumed = 0
            counterStart = time.time()
        else:
            consumed += number
        limiter.consumed(number)
        delay = limiter.delay()

        print(f"Consumed {number} bytes, sleeping for {delay} seconds")
        print(f"estimated speed {limiter._estimated_speed}")
        print(f"internal delay {limiter._delay}")
        print(f"Rate {consumed / (time.time() - counterStart)}")
        time.sleep(delay)
    
    while time.time() - start < 10:
        number = randint(1000, 2000) / 2
        time.sleep(0.01)
        consume(number)

    print(f"Consuming zeros")
    start = time.time()
    while time.time() - start < 10:
        number = 0
        time.sleep(0.01)
        consume(number)

    print(f"Consuming 100")
    start = time.time()
    while time.time() - start < 10:
        number = 10000
        time.sleep(0.01)
        consume(number)
    print(f"Consumed {consumed} bytes")
    print(f"estimated speed {limiter._estimated_speed}")
    print(f"Rate {consumed / (time.time() - counterStart)}")