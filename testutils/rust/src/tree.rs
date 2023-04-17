use std::cell::RefCell;
use std::collections::VecDeque;
use std::rc::Rc;

use serde::{Deserialize, Serialize, Serializer};
use serde::de::SeqAccess;
use serde::ser::SerializeSeq;

// LeetCode use `Option<Rc<RefCell<TreeNode>>>` for tree links, but `Option<Box<TreeNode>>` should be enough.
// https://github.com/pretzelhammer/rust-blog/blob/master/posts/learning-rust-in-2020.md#leetcode
type TreeLink = Option<Rc<RefCell<TreeNode>>>;

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct TreeNode {
    pub val: i32,
    pub left: TreeLink,
    pub right: TreeLink,
}

#[macro_export]
macro_rules! tree {
    () => {
        None
    };
    ($e:expr) => {
        Some(Rc::new(RefCell::new(TreeNode {
            val: $e,
            left: None,
            right: None,
        })))
    };
}

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct BinaryTree(TreeLink);

impl From<BinaryTree> for TreeLink {
    fn from(tree: BinaryTree) -> Self {
        tree.0
    }
}

impl Serialize for BinaryTree {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
        where
            S: Serializer,
    {
        let mut queue = VecDeque::new();
        let mut nodes = Vec::new();
        queue.push_back(self.0.clone());
        while let Some(node) = queue.pop_front() {
            nodes.push(node.clone());
            if let Some(ref node) = node {
                queue.push_back(node.borrow().left.clone());
                queue.push_back(node.borrow().right.clone());
            }
        }
        while nodes.len() > 0 && nodes.last().unwrap().is_none() {
            nodes.pop();
        }

        let mut seq = serializer.serialize_seq(None)?;
        for node in nodes {
            if let Some(ref node) = node {
                seq.serialize_element(&node.borrow().val.clone())?;
            } else {
                seq.serialize_element(&None::<i32>)?;
            }
        }
        seq.end()
    }
}

struct BinaryTreeVisitor;

impl<'de> serde::de::Visitor<'de> for BinaryTreeVisitor {
    type Value = BinaryTree;

    fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
        formatter.write_str("a list of optional integers")
    }

    fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
        where
            A: SeqAccess<'de>,
    {
        let mut nodes: Vec<TreeLink> = Vec::new();

        while let Some(val) = seq.next_element::<Option<i32>>()? {
            nodes.push(val.map(|v: i32| Rc::new(RefCell::new(TreeNode {
                val: v,
                left: None,
                right: None,
            }))));
        }

        let root = nodes[0].clone();
        let (mut i, mut j) = (0, 1);

        while j < nodes.len() {
            if let Some(ref current_node) = nodes[i] {
                current_node.borrow_mut().left = nodes[j].clone();
                j += 1;
                if j < nodes.len() {
                    current_node.borrow_mut().right = nodes[j].clone();
                    j += 1;
                }
            }
            i += 1;
        }

        Ok(BinaryTree(root))
    }
}

impl<'de> Deserialize<'de> for BinaryTree {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
        where
            D: serde::Deserializer<'de>,
    {
        deserializer.deserialize_seq(BinaryTreeVisitor)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tree_serialize() {
        let root = TreeNode {
            val: 1,
            left: Some(Rc::new(RefCell::new(TreeNode {
                val: 2,
                left: None,
                right: None,
            }))),
            right: Some(Rc::new(RefCell::new(TreeNode {
                val: 4,
                left: Some(Rc::new(RefCell::new(TreeNode {
                    val: 3,
                    left: None,
                    right: None,
                }))),
                right: None,
            }))),
        };
        let tree = BinaryTree(Some(Rc::new(RefCell::new(root))));
        let serialized = serde_json::to_string(&tree).unwrap();
        assert_eq!(serialized, "[1,2,4,null,null,3]");
    }
}
