import sys
from typing import List


def join_array(arr: List[str]) -> str:
    return "[" + ",".join(arr) + "]"


def read_line() -> str:
    return sys.stdin.readline().strip()
