use std::cell::RefCell;
use std::collections::VecDeque;
use std::ops::Deref;
use std::rc::Rc;

use serde::{Deserialize, Serialize, Serializer};
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

// struct BinaryTreeVisitor;
//
// impl<'de> serde::de::Visitor<'de> for BinaryTreeVisitor {
//     type Value = BinaryTree;
//
//     fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
//         formatter.write_str("a list of integers")
//     }
//
//     fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
//         where
//             A: serde::de::SeqAccess<'de>,
//     {
//         let mut nodes = Vec::new();
//         while let Some(val) = seq.next_element()? {
//             nodes.push(val);
//         }
//         let mut queue = VecDeque::new();
//         let mut root = None;
//         if nodes.len() > 0 {
//             root = Some(Rc::new(RefCell::new(TreeNode {
//                 val: nodes[0],
//                 left: None,
//                 right: None,
//             })));
//             queue.push_back(root.clone());
//         }
//         let mut i = 1;
//         while let Some(node) = queue.pop_front() {
//             if i < nodes.len() {
//                 let left = if nodes[i] == None {
//                     None
//                 } else {
//                     Some(Rc::new(RefCell::new(TreeNode {
//                         val: nodes[i].unwrap(),
//                         left: None,
//                         right: None,
//                     })))
//                 };
//                 node.as_ref().unwrap().borrow_mut().left = left.clone();
//                 if left.is_some() {
//                     queue.push_back(left);
//                 }
//                 i += 1;
//             }
//             if i < nodes.len() {
//                 let right = if nodes[i] == None {
//                     None
//                 } else {
//                     Some(Rc::new(RefCell::new(TreeNode {
//                         val: nodes[i].unwrap(),
//                         left: None,
//                         right: None,
//                     })))
//                 };
//                 node.as_ref().unwrap().borrow_mut().right = right.clone();
//                 if right.is_some() {
//                     queue.push_back(right);
//                 }
//                 i += 1;
//             }
//         }
//         Ok(BinaryTree(root))
//     }
// }
//
// impl<'de> Deserialize<'de> for BinaryTree {
//     fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
//         where
//             D: serde::Deserializer<'de>,
//     {
//         deserializer.deserialize_seq(BinaryTreeVisitor)
//     }
// }

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
