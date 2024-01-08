import json
from typing import Any, List

from . import ListNode, TreeNode


def split_array(s: str) -> List[str]:
    s = s.strip()
    if len(s) <= 1 or s[0] != "[" or s[-1] != "]":
        raise Exception("Invalid array: " + s)

    splits = json.loads(s)
    res = [json.dumps(split) for split in splits]
    return res


def serialize(val: Any, ty: str = None) -> str:
    if val is None:
        if ty is None:
            raise Exception("None value without type")
        if ty == "ListNode" or ty == "TreeNode":
            return "[]"
        return "null"
    elif isinstance(val, bool):
        return "true" if val else "false"
    elif isinstance(val, int):
        return str(val)
    elif isinstance(val, float):
        return str(val)
    elif isinstance(val, str):
        return '"' + val + '"'
    elif isinstance(val, list):
        return "[" + ",".join(serialize(v) for v in val) + "]"
    elif isinstance(val, (ListNode, TreeNode)):
        return val.serialize()
    else:
        raise Exception("Unknown type: " + str(type(val)))


def deserialize(ty: str, s: str) -> Any:
    if ty == "int":
        return int(s)
    elif ty == "float":
        return float(s)
    elif ty == "str":
        return s[1:-1]
    elif ty == "bool":
        return s == "true"
    elif ty.startswith("List["):
        arr = []
        for v in split_array(s):
            arr.append(deserialize(ty[5:-1], v))
        return arr
    elif ty == "ListNode":
        return ListNode.deserialize(s)
    elif ty == "TreeNode":
        return TreeNode.deserialize(s)
    else:
        raise Exception("Unknown type: " + ty)
