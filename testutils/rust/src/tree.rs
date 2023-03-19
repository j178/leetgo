use std::cell::RefCell;
use std::collections::VecDeque;
use std::ops::Deref;
use std::rc::Rc;

use serde::{Serialize, Serializer};

use serde::ser::SerializeSeq;

// LeetCode use `Option<Rc<RefCell<TreeNode>>>` for tree links, but `Option<Box<TreeNode>>` should be enough.
// https://github.com/pretzelhammer/rust-blog/blob/master/posts/learning-rust-in-2020.md#leetcode
pub type TreeLink = Option<Rc<RefCell<TreeNode>>>;

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct TreeNode {
    pub val: i32,
    pub left: TreeLink,
    pub right: TreeLink,
}

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
struct BinaryTree(TreeLink);

impl Deref for BinaryTree {
    type Target = TreeLink;

    fn deref(&self) -> &Self::Target {
        &self.0
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
