use std::collections::VecDeque;
use std::io::Write;

use serde::de::Visitor;
use serde::ser::SerializeSeq;
use serde::{Deserialize, Deserializer, Serialize, Serializer};

#[allow(dead_code)]
struct LeetCodeSerializer<W: Write> {
    writer: W,
}

#[derive(Debug)]
pub struct ListNode {
    pub val: i32,
    pub next: Option<Box<ListNode>>,
}

#[derive(Debug)]
struct LinkedList(Option<Box<ListNode>>);

impl Serialize for LinkedList {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        let mut seq = serializer.serialize_seq(None)?;
        let mut current = &self.0;
        while let Some(ref node) = current {
            seq.serialize_element(&node.val)?;
            current = &node.next;
        }
        seq.end()
    }
}

struct LinkedListVisitor;

impl<'de> Visitor<'de> for LinkedListVisitor {
    type Value = LinkedList;

    fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
        formatter.write_str("a list of integers")
    }

    fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
    where
        A: serde::de::SeqAccess<'de>,
    {
        let mut head = None;
        let mut current = &mut head;
        while let Some(val) = seq.next_element()? {
            let node = ListNode { val, next: None };
            *current = Some(Box::new(node));
            current = &mut current.as_mut().unwrap().next;
        }
        Ok(LinkedList(head))
    }
}

impl<'de> Deserialize<'de> for LinkedList {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        deserializer.deserialize_seq(LinkedListVisitor)
    }
}

#[derive(Debug)]
pub struct TreeNode {
    pub val: i32,
    pub left: Option<Box<TreeNode>>,
    pub right: Option<Box<TreeNode>>,
}

#[derive(Debug)]
struct BinaryTree(Option<Box<TreeNode>>);

impl Serialize for BinaryTree {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
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
    fn test_list_serialize() {
        let head = ListNode {
            val: 1,
            next: Some(Box::new(ListNode {
                val: 2,
                next: Some(Box::new(ListNode { val: 3, next: None })),
            })),
        };
        let list = LinkedList(Some(Box::new(head)));

        let serialized = serde_json::to_string(&list).unwrap();
        assert_eq!(serialized, "[1,2,3]");
    }

    #[test]
    fn test_list_deserialize() {
        let serialized = "[1,2,3]";
        let list: LinkedList = serde_json::from_str(serialized).unwrap();
        let mut current = &list.0;
        let mut i = 1;
        while let Some(ref node) = current {
            assert_eq!(node.val, i);
            current = &node.next;
            i += 1;
        }

        let serialized = "[]";
        let list: LinkedList = serde_json::from_str(serialized).unwrap();
        assert!(list.0.is_none());

        let serialized = "[true]";
        let list = serde_json::from_str::<LinkedList>(serialized);
        assert!(list.is_err());
    }

    #[test]
    fn test_tree_serialize() {
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
