import json
from typing import Optional


class TreeNode:
    def __init__(self, val: int, left: "TreeNode" = None, right: "TreeNode" = None):
        self.val = val
        self.left = left
        self.right = right

    @classmethod
    def deserialize(cls, s: str) -> Optional["TreeNode"]:
        res = json.loads(s)
        if not res:
            return None

        nodes = [TreeNode(val) if val is not None else None for val in res]
        root = nodes[0]

        j = 1
        for node in nodes:
            if node is not None:
                if j < len(res):
                    node.left = nodes[j]
                    j += 1
                if j < len(res):
                    node.right = nodes[j]
                    j += 1
            if j >= len(res):
                break

        return root

    def serialize(self) -> str:
        nodes = []
        queue = [self]

        while queue:
            t = queue.pop(0)
            nodes.append(t)

            if t is not None:
                queue.extend([t.left, t.right])

        while nodes and nodes[-1] is None:
            nodes.pop()

        arr = [node.val if node is not None else None for node in nodes]
        return json.dumps(arr)
