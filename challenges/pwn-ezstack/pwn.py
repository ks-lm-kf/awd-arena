#!/usr/bin/env python3
import sys
def vuln():
    buffer = b"A" * 64
    data = input("Enter your name: ")
    # Intentionally vulnerable - no bounds check
    buffer = data.encode()  # Simulated buffer overflow
    print(f"Hello {data}")
    if len(data) > 100:
        print("You overflowed the buffer!")
        with open("/flag.txt") as f:
            print(f.read())

if __name__ == "__main__":
    print("=== EZ Stack Overflow Challenge ===")
    print("Hint: Overflow the buffer to get the flag")
    vuln()

