import json
from typing import Optional


class ListNode:
    def __init__(self, val: int = 0, next: "ListNode" = None):
        self.val = val
        self.next = next

    @classmethod
    def deserialize(cls, s: str) -> Optional["ListNode"]:
        arr = json.loads(s)
        if not arr:
            return None
        root = ListNode(arr[0])
        node = root
        for v in arr[1:]:
            node.next = ListNode(v)
            node = node.next
        return root

    def serialize(self) -> str:
        s = ["["]
        node = self
        while node:
            s.append(str(node.val))
            node = node.next
            if node:
                s.append(",")
        s.append("]")
        return "".join(s)
