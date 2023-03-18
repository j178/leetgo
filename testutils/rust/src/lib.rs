use std::collections::VecDeque;
use std::io::Write;

use serde::{Deserialize, Deserializer, Serialize, Serializer};
use serde::ser::SerializeSeq;

#[allow(dead_code)]
struct LeetCodeSerializer<W: Write> {
    writer: W,
}

pub struct ListNode {
    pub val: i32,
    pub next: Option<Box<ListNode>>,
}

impl ListNode {
    pub fn new(val: i32) -> Self {
        Self {
            next: None,
            val,
        }
    }
}

struct LinkedList(Option<Box<ListNode>>);

impl Serialize for LinkedList {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error> where S: Serializer {
        let mut seq = serializer.serialize_seq(None)?;
        let mut current = &self.0;
        while let Some(ref node) = current {
            seq.serialize_element(&node.val)?;
            current = &node.next;
        }
        seq.end()
    }
}

// impl Deserialize<'_> for LinkedList {
//     fn deserialize<D>(deserializer: D) -> Result<Self, D::Error> where D: Deserializer<'_> {
//         let mut current = None;
//         for val in Vec::<i32>::deserialize(deserializer)? {
//             let mut node = ListNode::new(val);
//             node.next = current;
//             current = Some(Box::new(node));
//         }
//         Ok(LinkedList(current))
//     }
// }

pub struct TreeNode {
    pub val: i32,
    pub left: Option<Box<TreeNode>>,
    pub right: Option<Box<TreeNode>>,
}

struct BinaryTree(Option<Box<TreeNode>>);

impl Serialize for BinaryTree {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error> where S: Serializer {
        let mut queue: VecDeque<&Option<Box<TreeNode>>> = VecDeque::new();
        let mut nodes: Vec<&Option<Box<TreeNode>>> = Vec::new();
        queue.push_back(&self.0);
        while let Some(node) = queue.pop_front() {
            nodes.push(node);
            if let Some(ref node) = node {
                queue.push_back(&node.left);
                queue.push_back(&node.right);
            }
        }
        while nodes.len() > 0 && nodes.last().unwrap().is_none() {
            nodes.pop();
        }

        let mut seq = serializer.serialize_seq(None)?;
        for node in nodes {
            if let Some(ref node) = node {
                seq.serialize_element(&node.val)?;
            } else {
                seq.serialize_element(&None::<i32>)?;
            }
        }
        seq.end()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_list() {
        let head = ListNode {
            val: 1,
            next: Some(Box::new(ListNode {
                val: 2,
                next: Some(Box::new(ListNode {
                    val: 3,
                    next: None,
                })),
            })),
        };
        let list = LinkedList(Some(Box::new(head)));

        let serialized = serde_json::to_string(&list).unwrap();
        assert_eq!(serialized, "[1,2,3]");
    }

    #[test]
    fn test_tree() {
        let root = TreeNode {
            val: 1,
            left: Some(Box::new(TreeNode {
                val: 2,
                left: None,
                right: None,
            })),
            right: Some(Box::new(TreeNode {
                val: 4,
                left: Some(Box::new(TreeNode {
                    val: 3,
                    left: None,
                    right: None,
                })),
                right: None,
            })),
        };
        let tree = BinaryTree(Some(Box::new(root)));

        let serialized = serde_json::to_string(&tree).unwrap();
        assert_eq!(serialized, "[1,2,4,null,null,3]");
    }
}
